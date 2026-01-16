package main

import (
	"context"
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

// createTrackingMiddleware 创建工具调用追踪中间件（记录到后端数据库）
func createTrackingMiddleware(client *tools.APIClient) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
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
						"arguments": ctr.Params.Arguments,
						"duration":  duration.Milliseconds(),
						"success":   err == nil,
						"timestamp": time.Now().Format(time.RFC3339),
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
