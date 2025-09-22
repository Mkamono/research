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
	genkit.DefineTool(g, "chat", "Ask the user a question when you need clarification, additional information, or confirmation. Use this when you are uncertain about something, need user input to proceed, or want to verify your understanding before taking action. This will send a message to the user and wait for their reply.",
		func(ctx *ai.ToolContext, req app.ChatRequest) (app.ChatResponse, error) {
			return provider.Chat(ctx.Context, req)
		})
	genkit.DefineTool(g, "get_thread_history", "Get the conversation history of a specific thread. Use this to review previous messages in a conversation thread to understand context or see what has been discussed before.",
		func(ctx *ai.ToolContext, threadID string) (app.GetThreadHistoryResponse, error) {
			return provider.GetThreadHistory(ctx.Context, threadID)
		})
}
