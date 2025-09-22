package main

import (
	"context"
	"log"
	"net/http"
	"research/flow"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/firebase/genkit/go/plugins/mcp"
	"github.com/firebase/genkit/go/plugins/server"
)

func main() {
	ctx := context.Background()

	g := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
		genkit.WithDefaultModel("googleai/gemini-2.5-flash-lite"),
	)

	host, err := mcp.NewMCPHost(g, mcp.MCPHostOptions{})

	if err != nil {
		log.Fatal("Failed to create MCP host:", err)
	}

	for _, server := range servers {
		host.Connect(ctx, g, server.Name, server)
	}

	mcpTools, err := host.GetActiveTools(ctx, g)
	if err != nil {
		log.Fatal("Failed to get active tools:", err)
	}

	for _, tool := range mcpTools {
		genkit.RegisterAction(g, tool)
	}

	recipeGeneratorFlow := flow.RecipeGeneratorFlow(g)
	researchFlow := flow.ResearchFlow(g)
	simpleFlow := flow.SimpleFlow(g)

	// Start a server to serve the flow and keep the app running for the Developer UI
	mux := http.NewServeMux()
	mux.HandleFunc("POST /recipeGeneratorFlow", genkit.Handler(recipeGeneratorFlow))
	mux.HandleFunc("POST /researchFlow", genkit.Handler(researchFlow))
	mux.HandleFunc("POST /simpleFlow", genkit.Handler(simpleFlow))

	log.Println("Starting server on http://localhost:3400")
	log.Fatal(server.Start(ctx, "127.0.0.1:3400", mux))
}
