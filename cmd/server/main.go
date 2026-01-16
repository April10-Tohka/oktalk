package main

import (
	"fmt"
	"oktalk/internal/pkg/config"
	"oktalk/internal/pkg/log"
	"oktalk/internal/pkg/trace"
	"oktalk/internal/router"
	"oktalk/internal/servicecontext"

	"github.com/sirupsen/logrus"
)

func main() {
	// 1. 初始化配置
	conf := config.InitConfig()
	// 2. 日志配置
	log.InitLog(conf)
	// 3. trace配置
	shutdown := trace.InitOpenTelemetry()
	defer shutdown()
	// 4.
	svcctx := servicecontext.NewServiceContext(conf)

	// 5.初始化路由
	r := router.InitRouter(svcctx)

	// 6. 启动服务
	logrus.Infof("🔥 OKTalk Server 启动成功! 监听端口: %d", conf.Server.Port)
	if err := r.Run(fmt.Sprintf(":%d", conf.Server.Port)); err != nil {
		logrus.Fatalf("❌ Server 启动失败: %v", err)
	}
}
