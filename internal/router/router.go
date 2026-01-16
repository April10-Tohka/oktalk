package router

import (
	"oktalk/internal/controller"
	"oktalk/internal/middleware"
	"oktalk/internal/pkg/response"
	"oktalk/internal/service"
	"oktalk/internal/servicecontext"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/gin-gonic/gin"
)

func InitRouter(svcctx *servicecontext.ServiceContext) *gin.Engine {
	// 设置 Gin 模式 (debug/release)
	gin.SetMode(svcctx.Config.Server.Mode)

	r := gin.New()

	// 1. 挂载中间件
	r.Use(otelgin.Middleware(svcctx.Config.Server.ServerName))
	r.Use(middleware.TracingMiddleware())
	r.Use(middleware.RecoveryMiddleware()) // 防止程序崩溃
	r.Use(middleware.Cors())               // 跨域处理

	// 2. 初始化所有handler
	chatHandler := controller.NewChatHandler(service.NewChatService(svcctx))

	// 3. 基础路由
	r.GET("/ping", func(c *gin.Context) {
		response.SendJSON(c, 200, nil, "pong")

	})

	// 4.业务路由分组挂载
	apiV1 := r.Group("/api/v1")
	{
		// 调用各模块的注册函数，传入对应的 Handler
		RegisterChatRouter(apiV1, chatHandler)
		//RegisterEvalRouter(apiV1, evalHandler)
		//RegisterReportRouter(apiV1, reportHandler)
	}

	return r
}
