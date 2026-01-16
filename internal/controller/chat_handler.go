package controller

import (
	"fmt"
	"net/http"
	"oktalk/internal/service"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// VoiceChat 处理语音上传与 AI 对话
func (h *ChatHandler) VoiceChat(c *gin.Context) {
	ctx := c.Request.Context()

	// 2. 获取上传的文件
	file, err := c.FormFile("audio")
	if err != nil {
		logrus.WithContext(ctx).Errorf("❌ 获取上传文件失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "未检测到音频文件上传"})
		return
	}

	// 3. 确保临时目录存在
	tempDir := "storage/temp/audio"
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		_ = os.MkdirAll(tempDir, os.ModePerm)
	}

	// 4. 构建唯一文件名 (时间戳 + UUID)
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), uuid.New().String()[:8], filepath.Ext(file.Filename))
	savePath := filepath.Join(tempDir, filename)

	// 5. 保存文件到本地
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		logrus.WithContext(ctx).Errorf("❌ 保存文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统保存文件失败"})
		return
	}

	logrus.WithContext(ctx).Infof("✅ 语音文件上传成功: %s", savePath)

	reply, err := h.chatService.ProcessVoiceChat(ctx, savePath)
	logrus.WithContext(ctx).Info(reply)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 处理失败: " + err.Error()})
		return
	}

}
