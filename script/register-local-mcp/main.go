package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// MCP Server Names - 公開可能な定数
const (
	ServerAskMe = "ask-me"
)

const mcpTemplate = `package mcp

import (
	"github.com/firebase/genkit/go/plugins/mcp"
)

// MCP Server Names - 公開可能な定数
const (
{{range .Constants}}	{{.Name}} = "{{.Value}}"
{{end}})

var Servers = []mcp.MCPClientOptions{
{{range .Servers}}	{
		Name: {{.ConstantName}},
		Stdio: &mcp.StdioConfig{
			Command: "go",
			Args:    []string{"run", "{{.Path}}"},
		},
	},
{{end}}}
`

type MCPServer struct {
	Name         string
	Path         string
	ConstantName string
}

type Constant struct {
	Name  string
	Value string
}

type TemplateData struct {
	Servers   []MCPServer
	Constants []Constant
}

func main() {
	// サーバー名と定数名のマッピング
	serverConstantMap := map[string]string{
		ServerAskMe: "ServerAskMe",
	}

	// Find all MCP server directories
	mcpDir := "mcp"
	servers := []MCPServer{}
	constants := []Constant{}

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
				constantName, exists := serverConstantMap[serverName]
				if !exists {
					fmt.Printf("Warning: No constant defined for server '%s', skipping\n", serverName)
					continue
				}

				servers = append(servers, MCPServer{
					Name:         serverName,
					Path:         mainGoPath,
					ConstantName: constantName,
				})

				constants = append(constants, Constant{
					Name:  constantName,
					Value: serverName,
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

	outputFile := "mcp/local_mcp.go"
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", outputFile, err)
		os.Exit(1)
	}
	defer file.Close()

	data := TemplateData{
		Servers:   servers,
		Constants: constants,
	}
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
