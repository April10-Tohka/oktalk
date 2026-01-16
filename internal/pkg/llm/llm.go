package llm

import "context"

type LLMService interface {
	Chat(ctx context.Context, prompt string) (string, error)
}
