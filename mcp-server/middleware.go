package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/mingrenya/AI-Waf/mcp-server/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// createLoggingMiddleware 创建日志中间件（参考官方 examples/http/logging_middleware.go）
func createLoggingMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			start := time.Now()
			sessionID := req.GetSession().ID()

			// 记录请求
			log.Printf("[REQUEST] Session: %s | Method: %s", sessionID, method)

			// 详细记录工具调用
			if ctr, ok := req.(*mcp.CallToolRequest); ok {
				log.Printf("[TOOL CALL] Name: %s | Args: %v", ctr.Params.Name, ctr.Params.Arguments)
			}

			// 执行实际方法
			result, err := next(ctx, method, req)
			duration := time.Since(start)

			// 记录响应
			if err != nil {
				log.Printf("[RESPONSE] Session: %s | Method: %s | Status: ERROR | Duration: %v | Error: %v",
					sessionID, method, duration, err)
			} else {
				log.Printf("[RESPONSE] Session: %s | Method: %s | Status: OK | Duration: %v",
					sessionID, method, duration)

				// 记录工具结果详情
				if ctr, ok := result.(*mcp.CallToolResult); ok {
					log.Printf("[TOOL RESULT] IsError: %v | ContentCount: %d", ctr.IsError, len(ctr.Content))
				}
			}

			return result, err
		}
	}
}

// filterMCPMetadata 过滤MCP协议的元数据字段
// Claude Desktop等MCP客户端会在参数中添加parent_message_uuid等协议字段
// 这些字段不应该传递给业务API，需要过滤掉
func filterMCPMetadata(args map[string]interface{}) map[string]interface{} {
	// MCP协议相关的元数据字段列表
	mcpMetadataFields := []string{
		"parent_message_uuid",
		"_meta",
		"_clientInfo",
		"_requestId",
	}

	filtered := make(map[string]interface{})
	for k, v := range args {
		// 跳过元数据字段
		isMeta := false
		for _, metaField := range mcpMetadataFields {
			if k == metaField {
				isMeta = true
				log.Printf("[FILTER] Removing MCP metadata field: %s", k)
				break
			}
		}
		if !isMeta {
			filtered[k] = v
		}
	}
	return filtered
}

// extractParentMessageUUID 提取parent_message_uuid用于上下文追踪
func extractParentMessageUUID(args map[string]interface{}) string {
	if args == nil {
		return ""
	}
	if value, ok := args["parent_message_uuid"]; ok {
		switch v := value.(type) {
		case string:
			return v
		default:
			return ""
		}
	}
	return ""
}

// createTrackingMiddleware 创建工具调用追踪中间件（记录到后端数据库）
func createTrackingMiddleware(client *tools.APIClient) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			// 过滤MCP元数据字段（在执行之前）
			if ctr, ok := req.(*mcp.CallToolRequest); ok {
				if ctr.Params.Arguments != nil {
					var argMap map[string]interface{}
					if err := json.Unmarshal(ctr.Params.Arguments, &argMap); err == nil {
						// 提取上下文追踪字段并过滤掉MCP协议字段
						parentMessageUUID := extractParentMessageUUID(argMap)
						filtered := filterMCPMetadata(argMap)
						if parentMessageUUID != "" {
							ctx = context.WithValue(ctx, "parent_message_uuid", parentMessageUUID)
						}
						if filteredBytes, err := json.Marshal(filtered); err == nil {
							ctr.Params.Arguments = filteredBytes
						}
					}
				}
			}

			// 执行方法
			start := time.Now()
			result, err := next(ctx, method, req)
			duration := time.Since(start)

			// 仅记录工具调用
			if ctr, ok := req.(*mcp.CallToolRequest); ok {
				// 异步记录到后端数据库
				go func() {
					recordData := map[string]interface{}{
						"toolName":  ctr.Params.Name,
						"duration":  duration.Milliseconds(),
						"success":   err == nil,
						"timestamp": time.Now().Format(time.RFC3339),
					}
					if parentMessageUUID, ok := ctx.Value("parent_message_uuid").(string); ok && parentMessageUUID != "" {
						recordData["parent_message_uuid"] = parentMessageUUID
					}
					if err != nil {
						recordData["error"] = err.Error()
					}

					// 调用后端记录接口
					_, recordErr := client.Post("/api/v1/mcp/tool-calls/record", recordData)
					if recordErr != nil {
						log.Printf("[TRACKING] Warning: Failed to record tool call: %v", recordErr)
					} else {
						log.Printf("[TRACKING] Recorded: %s", ctr.Params.Name)
					}
				}()
			}

			return result, err
		}
	}
}
