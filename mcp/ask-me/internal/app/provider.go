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
	ThreadID *string `json:"thread_id,omitempty" desc:"CRITICAL: Thread ID to continue existing conversation. ALWAYS use the thread_id from previous chat responses for follow-up questions. Only omit for completely new, unrelated topics. Example: if asking follow-up questions about the same project/topic, use the same thread_id."`
}

type ChatResponse struct {
	Message  string `json:"message" desc:"The reply message received from the user"`
	ThreadID string `json:"thread_id" desc:"IMPORTANT: Save this thread_id and use it in all subsequent related questions. This maintains conversation context. Always include this thread_id when asking follow-up questions, clarifications, or discussing the same topic. Do NOT start new threads for related conversations."`
}

type GetThreadHistoryResponse struct {
	ThreadID string   `json:"thread_id" desc:"Thread ID of the conversation"`
	Messages []string `json:"messages" desc:"All messages in the thread in chronological order"`
}
