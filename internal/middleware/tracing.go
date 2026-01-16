// internal/middleware/tracing.go
package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"oktalk/internal/pkg/constants"
)

func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1.从 OpenTelemetry 的 Span 中获取 TraceID
		span := trace.SpanFromContext(c.Request.Context())
		traceID := span.SpanContext().TraceID().String()
		// 2. 存入标准 context (传递给 Service/DB 等)
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, constants.TraceIDKey, traceID)
		c.Request = c.Request.WithContext(ctx)
		// 3. 关键：存入 Gin 上下文，供 Handler 使用
		c.Set(constants.TraceIDKey, traceID)
		c.Next()
	}
}
