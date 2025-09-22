package main

import (
	"github.com/firebase/genkit/go/plugins/mcp"
)

var servers = []mcp.MCPClientOptions{
	{
		Name: "ask-me",
		Stdio: &mcp.StdioConfig{
			Command: "go",
			Args:    []string{"run", "mcp/ask-me/main.go"},
		},
	},
}
