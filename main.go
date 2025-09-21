package main

import (
	"context"
	"log"
	"net/http"

	"research/flow"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/firebase/genkit/go/plugins/server"
)

func main() {
	ctx := context.Background()

	// Initialize Genkit with the Google AI plugin
	g := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
		genkit.WithDefaultModel("googleai/gemini-2.5-flash-lite"),
	)

	// Create the recipe generator flow
	recipeGeneratorFlow := flow.RecipeGeneratorFlow(g)

	// Create the research flow
	researchFlow := flow.ResearchFlow(g)

	// Start a server to serve the flow and keep the app running for the Developer UI
	mux := http.NewServeMux()
	mux.HandleFunc("POST /recipeGeneratorFlow", genkit.Handler(recipeGeneratorFlow))
	mux.HandleFunc("POST /researchFlow", genkit.Handler(researchFlow))

	log.Println("Starting server on http://localhost:3400")
	log.Fatal(server.Start(ctx, "127.0.0.1:3400", mux))
}
