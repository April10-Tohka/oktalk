package service

import (
	"context"
	"oktalk/internal/pkg/asr"
	"oktalk/internal/pkg/llm"
	"oktalk/internal/pkg/tts"
	"oktalk/internal/servicecontext"

	"github.com/sirupsen/logrus"
)

type ChatService struct {
	svcctx     *servicecontext.ServiceContext
	asrService asr.ASRService
	llmService llm.LLMService
	ttsService tts.TTSService
}

func NewChatService(svcctx *servicecontext.ServiceContext) *ChatService {
	return &ChatService{
		asrService: asr.NewAliyunASR(&svcctx.Config.Aliyun),
		llmService: llm.NewQwenLLM(&svcctx.Config.Aliyun),
		ttsService: tts.NewAliyunTTS(&svcctx.Config.Aliyun),
	}
}

// ProcessVoiceChat 核心串联逻辑
func (s *ChatService) ProcessVoiceChat(ctx context.Context, audioPath string) (string, error) {
	// 1. ASR: 语音转文字
	// 注意：这里的 RecognizeOnce 需要实现你提供的示例代码中的 WebSocket 逻辑
	recognizedText, err := s.asrService.RecognizeOnce(ctx, audioPath)
	if err != nil {
		logrus.Errorf("ASR error: %v", err)
		return "", err
	}

	if recognizedText == "" {
		return "Sorry, I didn't hear anything clearly.", nil
	}

	logrus.Infof("🎙️ ASR Result: %s", recognizedText)

	// 2. LLM: 生成回复文本
	replyText, err := s.llmService.Chat(ctx, recognizedText)
	if err != nil {
		logrus.Errorf("LLM error: %v", err)
		return "", err
	}

	logrus.Infof("🤖 AI Reply: %s", replyText)

	// 3. 后续步骤：TTS (语音合成) 暂不在此处展示，通常在下一阶段实现
	return replyText, nil
}
