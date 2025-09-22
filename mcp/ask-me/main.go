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
	genkit.DefineTool(g, "chat", "Chat with a user",
		func(ctx *ai.ToolContext, req app.ChatRequest) (app.ChatResponse, error) {
			return provider.Chat(ctx.Context, req)
		})
	genkit.DefineTool(g, "get_thread_history", "Get thread history",
		func(ctx *ai.ToolContext, threadID string) (app.GetThreadHistoryResponse, error) {
			return provider.GetThreadHistory(ctx.Context, threadID)
		})
}
