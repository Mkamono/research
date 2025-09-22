package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const mcpTemplate = `package main

import (
	"github.com/firebase/genkit/go/plugins/mcp"
)

var servers = []mcp.MCPClientOptions{
{{range .Servers}}	{
		Name: "{{.Name}}",
		Stdio: &mcp.StdioConfig{
			Command: "go",
			Args:    []string{"run", "{{.Path}}"},
		},
	},
{{end}}}
`

type MCPServer struct {
	Name string
	Path string
}

type TemplateData struct {
	Servers []MCPServer
}

func main() {
	// Find all MCP server directories
	mcpDir := "mcp"
	servers := []MCPServer{}

	entries, err := os.ReadDir(mcpDir)
	if err != nil {
		fmt.Printf("Error reading mcp directory: %v\n", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			serverName := entry.Name()
			mainGoPath := filepath.Join(mcpDir, serverName, "main.go")

			// Check if main.go exists
			if _, err := os.Stat(mainGoPath); err == nil {
				servers = append(servers, MCPServer{
					Name: serverName,
					Path: mainGoPath,
				})
			}
		}
	}

	if len(servers) == 0 {
		fmt.Println("No MCP servers found in mcp/ directory")
		os.Exit(1)
	}

	// Generate local_mcp.go
	tmpl, err := template.New("mcp").Parse(mcpTemplate)
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		os.Exit(1)
	}

	outputFile := "local_mcp.go"
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", outputFile, err)
		os.Exit(1)
	}
	defer file.Close()

	data := TemplateData{Servers: servers}
	err = tmpl.Execute(file, data)
	if err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated %s with %d MCP servers:\n", outputFile, len(servers))
	for _, server := range servers {
		fmt.Printf("  - %s (%s)\n", server.Name, server.Path)
	}
}
