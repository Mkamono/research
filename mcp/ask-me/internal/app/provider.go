package app

import (
	"context"
	"time"
)

type ChatProvider interface {
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
	GetThreadHistory(ctx context.Context, threadID string) (GetThreadHistoryResponse, error)
}

type ChatRequest struct {
	Message  string        `json:"message" desc:"The message to send"`
	ThreadID *string       `json:"thread_id,omitempty" desc:"Optional thread ID to continue conversation (nil creates new thread)"`
	Timeout  time.Duration `json:"timeout" desc:"Timeout duration for waiting for reply"`
}

type ChatResponse struct {
	Message  string `json:"message" desc:"The reply message received"`
	ThreadID string `json:"thread_id" desc:"Thread ID for this conversation"`
}

type GetThreadHistoryResponse struct {
	ThreadID string   `json:"thread_id" desc:"Thread ID"`
	Messages []string `json:"messages" desc:"All messages in the thread"`
}
