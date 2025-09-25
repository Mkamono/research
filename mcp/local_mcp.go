package mcp

import (
	"github.com/firebase/genkit/go/plugins/mcp"
)

// MCP Server Names - 公開可能な定数
const (
	ServerAskMe = "ask-me"
)

var Servers = []mcp.MCPClientOptions{
	{
		Name: ServerAskMe,
		Stdio: &mcp.StdioConfig{
			Command: "go",
			Args:    []string{"run", "mcp/ask-me/main.go"},
		},
	},
}
