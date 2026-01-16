package tts

import "context"

type TTSService interface {
	// Synthesize 输入文本，输出生成的音频文件路径
	Synthesize(ctx context.Context, text string) ([]byte, error)
}
