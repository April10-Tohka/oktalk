package llm

import (
	"context"
	"oktalk/internal/pkg/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type QwenLLM struct {
	client openai.Client
	model  string
}

func NewQwenLLM(conf *config.AliyunConfig) *QwenLLM {
	client := openai.NewClient(
		option.WithAPIKey(conf.DASHSCOPE_API_KEY),
		option.WithBaseURL(conf.LLM.BaseURL),
	)
	return &QwenLLM{
		client: client,
		model:  conf.LLM.Model,
	}
}

func (q *QwenLLM) Chat(ctx context.Context, prompt string) (string, error) {
	// 这里的 System Prompt 是为了体现 PRD 中“耐心英语老师”的角色设定
	chatCompletion, err := q.client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage("You are a patient and friendly English teacher for kids aged 6-12. Use simple words and keep responses short."),
				openai.UserMessage(prompt),
			},
			Model: q.model,
		},
	)
	if err != nil {
		return "", err
	}
	return chatCompletion.Choices[0].Message.Content, nil
}
