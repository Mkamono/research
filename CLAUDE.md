# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based research application using Firebase Genkit that implements a multi-agent AI system for research and recipe generation. The application provides structured workflows through two main flows: `RecipeGeneratorFlow` and `ResearchFlow`.

## Architecture

### Main Components

- **main.go**: Entry point that initializes Genkit with Google AI plugin, sets up HTTP handlers for flows, and starts the server on port 3400
- **flow/recipe.go**: Recipe generation flow that takes ingredient and dietary restrictions as input and generates structured recipe data
- **flow/research.go**: Multi-agent research flow implementing a 3-agent system:
  - Planning Agent: Creates research plan with questions and approach
  - Data Collection Agent: Gathers findings and sources based on the plan
  - Synthesis Agent: Creates comprehensive research report (outputs in Japanese)

### Key Features

- Uses Firebase Genkit Go SDK with Google AI (Gemini 2.5 Flash Lite) as the default model
- Structured data generation with JSON schemas for input/output validation
- HTTP endpoints for flow execution
- Streaming research flow available for real-time progress updates
- Multi-agent orchestration for complex research tasks

## Development Commands

All commands are managed through `mise` (configured in `mise.toml`):

```bash
# Start development server with Genkit UI
mise run dev

# Kill running processes (primary method)
mise run kill2

# Alternative kill method
mise run kill
```

The development server:
- Runs `genkit start -- go run .`
- Provides Genkit Developer UI for testing flows
- Serves HTTP endpoints at http://localhost:3400

## API Endpoints

- `POST /recipeGeneratorFlow`: Recipe generation
- `POST /researchFlow`: Research report generation

## Dependencies

- Go 1.24.5
- Firebase Genkit Go SDK
- Google AI plugin for Gemini models
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

The application requires the Genkit CLI to be available (`npm:genkit-cli` is installed via mise).