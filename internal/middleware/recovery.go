package middleware

import (
	"fmt"
	"net"
	"net/http/httputil"
	"oktalk/internal/pkg/constants"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"oktalk/internal/pkg/response"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 1. 获取 TraceID
				traceID, _ := c.Get(constants.TraceIDKey)
				ctx := c.Request.Context()
				// 2. 检查是否是断开连接导致的 panic (比如客户端主动关闭连接)
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				// 获取请求详情用于日志
				httpRequest, _ := httputil.DumpRequest(c.Request, false)

				if brokenPipe {
					logrus.WithContext(ctx).Printf("[Broken Pipe] trace_id: %s, error: %v\n%s", traceID, err, string(httpRequest))
					c.Error(err.(error))
					c.Abort()
					return
				}

				// 3. 打印详细的堆栈信息到控制台/日志
				// 在实际毕设中，这能帮你迅速定位代码哪一行崩了
				stack := debug.Stack()
				logrus.WithContext(ctx).Printf("[Panic Recover] trace_id: %s, err: %v\n%s\n%s",
					traceID, err, string(httpRequest), string(stack))

				// 4. 关键点：中断请求并返回带 TraceID 的 JSON
				// 使用 AbortWithStatusJSON 确保后续的 Handler 不再执行
				response.SendJSON(c, 500,
					nil, fmt.Sprintf("服务器内部错误: %v", err))

				c.Abort()
			}
		}()
		c.Next()
	}
}
