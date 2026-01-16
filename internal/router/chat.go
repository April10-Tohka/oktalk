package router

import (
	"oktalk/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterChatRouter 注册聊天模块路由
func RegisterChatRouter(v1 *gin.RouterGroup, handler *controller.ChatHandler) {
	chat := v1.Group("/chat")
	{
		chat.POST("/voice", handler.VoiceChat) // 映射到结构体方法
	}
}
