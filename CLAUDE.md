# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based application using Firebase Genkit that provides AI-powered workflows for recipe generation and simple text processing. The application currently has two main flows: `RecipeGeneratorFlow` and `SimpleFlow`, with MCP (Model Context Protocol) integration for external tool access.

## Architecture

### Main Components

- **main.go**: Entry point that initializes Genkit with Google AI plugin, sets up MCP (Model Context Protocol) host for external tools, and starts HTTP server on port 3400
- **flow/recipe.go**: Recipe generation flow that takes ingredient and dietary restrictions as input and generates structured recipe data
- **flow/simple.go**: Basic text generation flow for simple AI interactions with MCP tool access
- **local_mcp.go**: Configuration for MCP servers that provide external tools (ask-me server)
- **mcp/ask-me/**: MCP server that provides chat functionality with Slack integration for user interaction

### Key Features

- Uses Firebase Genkit Go SDK with Google AI (Gemini 2.5 Flash Lite) as the default model
- MCP (Model Context Protocol) integration for external tool access
- Structured data generation with JSON schemas for input/output validation
- HTTP endpoints for flow execution
- Slack integration for interactive user communication through MCP tools

### MCP Integration

The application connects to local MCP servers defined in `local_mcp.go`:
- **ask-me**: Interactive chat server (`mcp/ask-me/main.go`) that provides:
  - `chat`: Tool for asking users questions and getting responses via Slack
  - `get_thread_history`: Tool for retrieving conversation history from threads

These servers provide tools that are registered as Genkit actions and available to flows that support MCP tools.

## Development Commands

All commands are managed through `mise` (configured in `mise.toml`):

```bash
# Start development server with Genkit UI
mise run dev

# Kill research-specific ports (4033, 4000, 3400)
mise run kill-research-ports

# Register MCP servers
mise run register-mcp
```

The development server:
- Runs `genkit start -- go run .`
- Provides Genkit Developer UI for testing flows
- Serves HTTP endpoints at http://localhost:3400

## API Endpoints

- `POST /recipeGeneratorFlow`: Recipe generation
- `POST /simpleFlow`: Basic text generation with MCP tool support

## Input/Output Schemas

### Recipe Flow
- Input: `RecipeInput` with ingredient and dietary restrictions
- Output: `Recipe` with structured recipe data including ingredients, instructions, and timing

### Simple Flow
- Input: `SimpleInput` with text input
- Output: Plain text response
- Has access to MCP tools for interactive user communication

## Dependencies

- Go 1.24.5
- Firebase Genkit Go SDK
- Google AI plugin for Gemini models
- MCP Go library for external tool integration
- GEMINI_API_KEY environment variable (configured in .env.local)
- SLACK_OAUTH_TOKEN and SLACK_CHANNEL environment variables for MCP integration

## Build and Run

```bash
# Development with Genkit UI
mise run dev

# Direct execution
go run .

# Clean up processes
mise run kill-research-ports
```

The application requires:
- Genkit CLI (`npm:genkit-cli` installed via mise)
- Valid GEMINI_API_KEY environment variable in .env.local
- SLACK_OAUTH_TOKEN and SLACK_CHANNEL for MCP chat functionality
- Go 1.24.5 or later

## Testing and Development

The project uses Firebase Genkit's built-in development tools. The Genkit Developer UI provides:
- Flow testing interface
- Request/response inspection
- Performance monitoring
- MCP tool interaction testing