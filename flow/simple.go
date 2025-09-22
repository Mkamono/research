// host, err := mcp.NewMCPHost(g, mcp.MCPHostOptions{})
// if err != nil {
// 	log.Fatal("Failed to create MCP host:", err)
// }

// host.Connect(ctx, g, "local", mcp.MCPClientOptions{
// 	Name: "everything-server",
// 	Stdio: &mcp.StdioConfig{
// 		Command: "npx",
// 		Args:    []string{"-y", "@modelcontextprotocol/server-everything"},
// 	},
// })

// tools, err := host.GetActiveTools(ctx, g)
// if err != nil {
// 	log.Fatal("Failed to get active tools:", err)
// }

package flow

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

// Define input schema
type SimpleInput struct {
	Input string `json:"input" jsonschema:"description=User input text"`
}

func SimpleFlow(g *genkit.Genkit) *core.Flow[*SimpleInput, string, struct{}] {
	// Define a simple flow that sends input to AI and returns response

	simpleFlow := genkit.DefineFlow(g, "simpleFlow", func(ctx context.Context, input *SimpleInput) (string, error) {

		// Generate response using AI
		response, err := genkit.GenerateText(ctx, g,
			ai.WithPrompt(input.Input),
		)
		if err != nil {
			return "", fmt.Errorf("failed to generate response: %w", err)
		}

		return response, nil
	})

	return simpleFlow
}
