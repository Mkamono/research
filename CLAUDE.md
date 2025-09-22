# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based research application using Firebase Genkit that implements a multi-agent AI system for research and recipe generation. The application provides structured workflows through three main flows: `RecipeGeneratorFlow`, `ResearchFlow`, and `SimpleFlow`.

## Architecture

### Main Components

- **main.go**: Entry point that initializes Genkit with Google AI plugin, sets up MCP (Model Context Protocol) host for external tools, and starts HTTP server on port 3400
- **flow/recipe.go**: Recipe generation flow that takes ingredient and dietary restrictions as input and generates structured recipe data
- **flow/research.go**: Multi-agent research flow implementing a 3-agent system:
  - Planning Agent: Creates research plan with questions and approach
  - Data Collection Agent: Gathers findings and sources based on the plan
  - Synthesis Agent: Creates comprehensive research report (outputs in Japanese)
- **flow/simple.go**: Basic text generation flow for simple AI interactions
- **servers.go**: Configuration for MCP servers that provide external tools (add, add2 servers)

### Key Features

- Uses Firebase Genkit Go SDK with Google AI (Gemini 2.5 Flash Lite) as the default model
- MCP (Model Context Protocol) integration for external tool access
- Structured data generation with JSON schemas for input/output validation
- HTTP endpoints for flow execution
- Streaming research flow available for real-time progress updates
- Multi-agent orchestration for complex research tasks with optional user input providers

### MCP Integration

The application connects to local MCP servers defined in `servers.go`:
- **add**: Basic addition server (`mcp/add/main.go`)
- **add2**: Extended addition server (`mcp/add2/main.go`)

These servers provide tools that are registered as Genkit actions and available to all flows.

## Development Commands

All commands are managed through `mise` (configured in `mise.toml`):

```bash
# Start development server with Genkit UI
mise run dev

# Kill running processes (primary method)
mise run kill2

# Alternative kill method
mise run kill

# Check running ports
mise run ports

# Kill specific ports
mise run kill-ports
mise run kill-research-ports
```

The development server:
- Runs `genkit start -- go run .`
- Provides Genkit Developer UI for testing flows
- Serves HTTP endpoints at http://localhost:3400

## API Endpoints

- `POST /recipeGeneratorFlow`: Recipe generation
- `POST /researchFlow`: Research report generation (3-agent workflow)
- `POST /simpleFlow`: Basic text generation

## Input/Output Schemas

### Recipe Flow
- Input: `RecipeInput` with ingredient and dietary restrictions
- Output: `Recipe` with structured recipe data including ingredients, instructions, and timing

### Research Flow
- Input: `ResearchInput` with topic, scope, depth, sources, context, purpose, and timeframe
- Output: `ResearchResult` with comprehensive research report including executive summary, key findings, recommendations, and agent trace

### Simple Flow
- Input: `SimpleInput` with text input
- Output: Plain text response

## Dependencies

- Go 1.24.5
- Firebase Genkit Go SDK
- Google AI plugin for Gemini models
- MCP Go library for external tool integration
- GEMINI_API_KEY environment variable (configured in mise.toml)

## Build and Run

```bash
# Development with Genkit UI
mise run dev

# Direct execution
go run .

# Clean up processes
mise run kill2
```

The application requires:
- Genkit CLI (`npm:genkit-cli` installed via mise)
- Valid GEMINI_API_KEY environment variable
- Go 1.24.5 or later

## Testing and Development

The project uses Firebase Genkit's built-in development tools. The Genkit Developer UI provides:
- Flow testing interface
- Request/response inspection
- Performance monitoring
- Agent trace visualization for multi-agent workflows