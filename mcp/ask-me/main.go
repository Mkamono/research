package main

import (
	"context"
	"log"
	"os"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/mcp"

	"research/mcp/ask-me/internal/app"
	"research/mcp/ask-me/internal/provider/slack"

	"github.com/firebase/genkit/go/ai"
)

func main() {
	ctx := context.Background()
	g := genkit.Init(ctx)

	chatProvider := slack.NewChatProvider(
		os.Getenv("SLACK_OAUTH_TOKEN"),
		os.Getenv("SLACK_CHANNEL"),
	)

	registerTools(g, chatProvider)

	server := mcp.NewMCPServer(g, mcp.MCPServerOptions{
		Name:    "ask-me",
		Version: "1.0.0",
	})

	if err := server.ServeStdio(); err != nil {
		log.Fatal(err)
	}
}

func registerTools(g *genkit.Genkit, provider app.ChatProvider) {
	genkit.DefineTool(g, "chat", "Ask the user a question when you need clarification, additional information, or confirmation. IMPORTANT: If this is a follow-up to a previous conversation, ALWAYS include the thread_id from the previous response to continue in the same thread. Only leave thread_id null for completely new topics. This maintains conversation context and keeps related discussions together.",
		func(ctx *ai.ToolContext, req app.ChatRequest) (app.ChatResponse, error) {
			return provider.Chat(ctx.Context, req)
		})
	genkit.DefineTool(g, "get_thread_history", "Get the conversation history of a specific thread. Use this to review previous messages in a conversation thread to understand context or see what has been discussed before.",
		func(ctx *ai.ToolContext, threadID string) (app.GetThreadHistoryResponse, error) {
			return provider.GetThreadHistory(ctx.Context, threadID)
		})
}
