package app

import (
	"context"
)

type ChatProvider interface {
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
	GetThreadHistory(ctx context.Context, threadID string) (GetThreadHistoryResponse, error)
}

type ChatRequest struct {
	Message  string  `json:"message" desc:"The message to send to the user"`
	ThreadID *string `json:"thread_id,omitempty" desc:"Optional thread ID to continue an existing conversation. Use the same thread ID for related questions or follow-up discussions. Leave nil to start a new conversation thread."`
}

type ChatResponse struct {
	Message  string `json:"message" desc:"The reply message received from the user"`
	ThreadID string `json:"thread_id" desc:"Thread ID for this conversation. Save this to continue the conversation in the same thread for related topics or follow-up questions."`
}

type GetThreadHistoryResponse struct {
	ThreadID string   `json:"thread_id" desc:"Thread ID of the conversation"`
	Messages []string `json:"messages" desc:"All messages in the thread in chronological order"`
}
