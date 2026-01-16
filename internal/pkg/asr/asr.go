package asr

import "context"

// Result ASR 识别结果
type Result struct {
	Text     string // 最终识别的文本
	IsFinal  bool   // 是否是最终结果
	Duration int    // 音频时长（毫秒）
}

// ASRService ASR 服务接口
type ASRService interface {
	// Start 开启一个识别会话（返回用于发送音频流的管道和接收结果的管道）
	// 这是处理流式语音的核心逻辑
	// RecognizeStream(ctx context.Context) (dataChan chan []byte, errChan chan error, resChan chan Result, err error)

	// RecognizeOnce 处理已经录好的完整文件（PRD 中的简单模式）
	RecognizeOnce(ctx context.Context, audioPath string) (string, error)
}
