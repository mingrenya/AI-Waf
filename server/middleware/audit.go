package middleware

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mingrenya/AI-Waf/server/config"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// AuditLog 审计日志结构
type AuditLog struct {
	RequestID    string                 `bson:"request_id" json:"request_id"`
	UserID       string                 `bson:"user_id" json:"user_id"`
	Username     string                 `bson:"username" json:"username"`
	UserRole     string                 `bson:"user_role" json:"user_role"`
	Method       string                 `bson:"method" json:"method"`
	Path         string                 `bson:"path" json:"path"`
	ClientIP     string                 `bson:"client_ip" json:"client_ip"`
	UserAgent    string                 `bson:"user_agent" json:"user_agent"`
	RequestBody  string                 `bson:"request_body,omitempty" json:"request_body,omitempty"`
	ResponseCode int                    `bson:"response_code" json:"response_code"`
	Duration     int64                  `bson:"duration" json:"duration"` // 毫秒
	Timestamp    time.Time              `bson:"timestamp" json:"timestamp"`
	Action       string                 `bson:"action" json:"action"`
	Resource     string                 `bson:"resource,omitempty" json:"resource,omitempty"`
	Changes      map[string]interface{} `bson:"changes,omitempty" json:"changes,omitempty"`
	Success      bool                   `bson:"success" json:"success"`
	ErrorMessage string                 `bson:"error_message,omitempty" json:"error_message,omitempty"`
}

// SecurityAudit 安全审计中间件 - 记录所有敏感操作
// 需要在router中传入db实例
func SecurityAudit(db *mongo.Database) gin.HandlerFunc {
	log := config.Logger
	collection := db.Collection("audit_logs")

	// 需要审计的路径前缀
	auditPaths := []string{
		"/api/v1/users",
		"/api/v1/site",
		"/api/v1/certificate",
		"/api/v1/ip-groups",
		"/api/v1/micro-rules",
		"/api/v1/rules",
		"/api/v1/config",
		"/api/v1/blocked-ips",
		"/api/v1/alerts",
		"/api/v1/ai-analyzer",
		"/api/v1/mcp",
	}

	// 需要审计的HTTP方法
	auditMethods := map[string]bool{
		"POST":   true,
		"PUT":    true,
		"PATCH":  true,
		"DELETE": true,
	}

	return func(c *gin.Context) {
		// 检查是否需要审计
		needAudit := false
		path := c.Request.URL.Path
		method := c.Request.Method

		// 检查方法和路径
		if auditMethods[method] {
			for _, prefix := range auditPaths {
				if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
					needAudit = true
					break
				}
			}
		}

		if !needAudit {
			c.Next()
			return
		}

		// 记录开始时间
		startTime := time.Now()

		// 读取请求体（如果有）
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // 重新设置body以便后续使用
			if len(bodyBytes) > 0 && len(bodyBytes) < 10000 {         // 限制大小
				requestBody = string(bodyBytes)
			}
		}

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(startTime).Milliseconds()

		// 构建审计日志
		auditLog := AuditLog{
			RequestID:    c.GetString("requestID"),
			Method:       method,
			Path:         path,
			ClientIP:     c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			RequestBody:  requestBody,
			ResponseCode: c.Writer.Status(),
			Duration:     duration,
			Timestamp:    startTime,
			Success:      c.Writer.Status() < 400,
		}

		// 获取用户信息
		if userID, exists := c.Get("userID"); exists {
			auditLog.UserID = userID.(string)
		}
		if username, exists := c.Get("username"); exists {
			auditLog.Username = username.(string)
		}
		if userRole, exists := c.Get("userRole"); exists {
			auditLog.UserRole = userRole.(string)
		}

		// 解析操作和资源
		auditLog.Action = determineAction(method, path)
		auditLog.Resource = extractResource(path)

		// 获取错误信息（如果有）
		if len(c.Errors) > 0 {
			auditLog.ErrorMessage = c.Errors.String()
		}

		// 异步保存审计日志到MongoDB
		// 使用独立 context，避免请求结束后原 context 被取消导致写入失败
		go func(log AuditLog) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if _, err := collection.InsertOne(ctx, log); err != nil {
				config.Logger.Error().Err(err).Msg("Failed to save audit log")
			}
		}(auditLog)

		// 同时输出到日志
		logEvent := log.Info()
		if !auditLog.Success {
			logEvent = log.Warn()
		}

		logEvent.
			Str("request_id", auditLog.RequestID).
			Str("user_id", auditLog.UserID).
			Str("username", auditLog.Username).
			Str("action", auditLog.Action).
			Str("resource", auditLog.Resource).
			Str("method", method).
			Str("path", path).
			Int("status", auditLog.ResponseCode).
			Int64("duration_ms", duration).
			Msg("Security Audit")
	}
}

// determineAction 根据HTTP方法和路径确定操作类型
func determineAction(method, _ string) string {
	switch method {
	case "POST":
		return "CREATE"
	case "PUT", "PATCH":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	case "GET":
		return "READ"
	default:
		return method
	}
}

// extractResource 从路径中提取资源类型
func extractResource(path string) string {
	parts := splitAndTrim(path, "/")
	if len(parts) >= 2 {
		return parts[1] // 返回 /api/{resource} 中的 resource
	}
	return ""
}

// AuditQueryResult 查询审计日志的结果
type AuditQueryResult struct {
	Logs  []AuditLog `json:"logs"`
	Total int64      `json:"total"`
}

// QueryAuditLogs 查询审计日志（用于管理界面）
func QueryAuditLogs(db *mongo.Database, filter bson.M, skip, limit int64) (*AuditQueryResult, error) {
	collection := db.Collection("audit_logs")
	ctx := context.Background()

	// 计数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	// 查询（应用分页和排序）
	findOpts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "timestamp", Value: -1}})
	cursor, err := collection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []AuditLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}

	return &AuditQueryResult{
		Logs:  logs,
		Total: total,
	}, nil
}

// CleanupOldAuditLogs 清理旧的审计日志（保留最近N天）
func CleanupOldAuditLogs(db *mongo.Database, daysToKeep int) error {
	collection := db.Collection("audit_logs")
	ctx := context.Background()

	cutoffTime := time.Now().AddDate(0, 0, -daysToKeep)
	filter := bson.M{
		"timestamp": bson.M{"$lt": cutoffTime},
	}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}

	config.Logger.Info().
		Int64("deleted_count", result.DeletedCount).
		Int("days_kept", daysToKeep).
		Msg("Cleaned up old audit logs")

	return nil
}
