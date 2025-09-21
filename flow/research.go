package flow

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

// UserInputProvider defines an interface for getting user input during research
type UserInputProvider interface {
	AskUser(question, context string) (string, error)
}

// DefaultUserInputProvider implements UserInputProvider with no-op
type DefaultUserInputProvider struct{}

func (d *DefaultUserInputProvider) AskUser(question, context string) (string, error) {
	// Default implementation returns empty response - no user interaction
	return "", nil
}

// MCPUserInputProvider implements UserInputProvider for MCP integration
type MCPUserInputProvider struct {
	askUserFunc func(question, context string) (string, error)
}

func NewMCPUserInputProvider(askFunc func(question, context string) (string, error)) *MCPUserInputProvider {
	return &MCPUserInputProvider{askUserFunc: askFunc}
}

func (m *MCPUserInputProvider) AskUser(question, context string) (string, error) {
	if m.askUserFunc != nil {
		return m.askUserFunc(question, context)
	}
	return "", nil
}

// Define input schema
type ResearchInput struct {
	Topic     string `json:"topic" jsonschema:"description=Research topic or question,required"`
	Scope     string `json:"scope,omitempty" jsonschema:"description=Research scope (broad, narrow, specific area)"`
	Depth     string `json:"depth,omitempty" jsonschema:"description=Research depth (overview, detailed, comprehensive)"`
	Sources   string `json:"sources,omitempty" jsonschema:"description=Preferred source types (academic, industry, news, etc.)"`
	Context   string `json:"context,omitempty" jsonschema:"description=Additional context or background information"`
	Purpose   string `json:"purpose,omitempty" jsonschema:"description=Purpose or intended use of the research"`
	Timeframe string `json:"timeframe,omitempty" jsonschema:"description=Time period to focus on (recent, historical, etc.)"`
}

// Simplified data structures for 3-agent workflow
type ResearchPlan struct {
	ResearchQuestions []string `json:"researchQuestions"`
	Approach          string   `json:"approach"`
	KeyFocus          []string `json:"keyFocus"`
}

type ResearchData struct {
	Findings []string `json:"findings"`
	Sources  []string `json:"sources"`
	Insights []string `json:"insights"`
}

// Define output schema
type ResearchResult struct {
	Title            string   `json:"title"`
	ExecutiveSummary string   `json:"executiveSummary"`
	KeyFindings      []string `json:"keyFindings"`
	MainPoints       []string `json:"mainPoints"`
	Recommendations  []string `json:"recommendations,omitempty"`
	NextSteps        []string `json:"nextSteps,omitempty"`
	Sources          []string `json:"sources,omitempty"`
	Limitations      []string `json:"limitations,omitempty"`
	AgentTrace       []string `json:"agentTrace,omitempty"`
}

// Agent 1: Research Planning Agent
func planningAgent(ctx context.Context, g *genkit.Genkit, input *ResearchInput, userInput UserInputProvider) (*ResearchPlan, error) {
	// Ask user for additional planning guidance if needed
	var additionalGuidance string
	if userInput != nil {
		guidance, err := userInput.AskUser(
			"Do you have any specific research questions or focus areas you'd like to emphasize for this topic?",
			fmt.Sprintf("Topic: %s, Scope: %s", input.Topic, getValueOrDefault(input.Scope, "comprehensive")),
		)
		if err == nil && guidance != "" {
			additionalGuidance = guidance
		}
	}

	prompt := fmt.Sprintf(`You are a Research Planning Agent. Create a focused research plan:

Topic: %s
Scope: %s
Depth: %s
Context: %s
Additional User Guidance: %s

Create a clear research plan with:
1. 3-5 specific research questions
2. Research approach/methodology
3. Key focus areas to investigate

Keep it concise and actionable.`,
		input.Topic,
		getValueOrDefault(input.Scope, "comprehensive"),
		getValueOrDefault(input.Depth, "detailed"),
		getValueOrDefault(input.Context, "general research"),
		additionalGuidance)

	plan, _, err := genkit.GenerateData[ResearchPlan](ctx, g, ai.WithPrompt(prompt))
	return plan, err
}

// Agent 2: Research Data Collection Agent
func dataCollectionAgent(ctx context.Context, g *genkit.Genkit, input *ResearchInput, plan *ResearchPlan, userInput UserInputProvider) (*ResearchData, error) {
	// Ask user for specific data requirements or sources
	var dataGuidance string
	if userInput != nil {
		guidance, err := userInput.AskUser(
			"Are there any specific sources, databases, or types of information you'd like me to prioritize for this research?",
			fmt.Sprintf("Research Questions: %v", plan.ResearchQuestions),
		)
		if err == nil && guidance != "" {
			dataGuidance = guidance
		}
	}

	prompt := fmt.Sprintf(`You are a Research Data Collection Agent. Based on the research plan, gather relevant information:

Topic: %s
Research Questions: %v
Approach: %s
Key Focus: %v
User Data Guidance: %s

Your task:
1. Collect key findings for each research question
2. Identify credible sources
3. Extract important insights

Provide comprehensive findings that address the research questions.`,
		input.Topic, plan.ResearchQuestions, plan.Approach, plan.KeyFocus, dataGuidance)

	data, _, err := genkit.GenerateData[ResearchData](ctx, g, ai.WithPrompt(prompt))
	return data, err
}

// Agent 3: Research Synthesis Agent
func synthesisAgent(ctx context.Context, g *genkit.Genkit, input *ResearchInput, plan *ResearchPlan, data *ResearchData, userInput UserInputProvider) (*ResearchResult, error) {
	// Ask user about report preferences
	var reportGuidance string
	if userInput != nil {
		guidance, err := userInput.AskUser(
			"What is the intended audience or specific focus for this research report? Any particular format or emphasis you'd prefer?",
			fmt.Sprintf("Topic: %s, Purpose: %s", input.Topic, getValueOrDefault(input.Purpose, "general research")),
		)
		if err == nil && guidance != "" {
			reportGuidance = guidance
		}
	}

	prompt := fmt.Sprintf(`You are a Research Synthesis Agent. Create a comprehensive research report:

Topic: %s
Research Plan: %v
Findings: %v
Sources: %v
Insights: %v
User Report Guidance: %s

Create a professional research report with:
1. Clear title
2. Executive summary
3. Key findings
4. Main points
5. Recommendations
6. Next steps
7. Sources
8. Limitations

最後に、日本語で出力してください。`,
		input.Topic, plan, data.Findings, data.Sources, data.Insights, reportGuidance)

	result, _, err := genkit.GenerateData[ResearchResult](ctx, g, ai.WithPrompt(prompt))
	return result, err
}

// Helper function for default values
func getValueOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// Simple 3-Agent Research Flow
func ResearchFlow(g *genkit.Genkit) *core.Flow[*ResearchInput, *ResearchResult, struct{}] {
	return ResearchFlowWithUserInput(g, &DefaultUserInputProvider{})
}

// Research Flow with User Input Provider
func ResearchFlowWithUserInput(g *genkit.Genkit, userInput UserInputProvider) *core.Flow[*ResearchInput, *ResearchResult, struct{}] {
	researchFlow := genkit.DefineFlow(g, "researchFlow", func(ctx context.Context, input *ResearchInput) (*ResearchResult, error) {
		var agentTrace []string

		// Step 1: Planning Agent
		agentTrace = append(agentTrace, "Starting planning agent...")
		plan, err := planningAgent(ctx, g, input, userInput)
		if err != nil {
			return nil, fmt.Errorf("planning agent failed: %w", err)
		}
		agentTrace = append(agentTrace, "Planning agent completed successfully")

		// Step 2: Data Collection Agent
		agentTrace = append(agentTrace, "Starting data collection agent...")
		data, err := dataCollectionAgent(ctx, g, input, plan, userInput)
		if err != nil {
			return nil, fmt.Errorf("data collection agent failed: %w", err)
		}
		agentTrace = append(agentTrace, "Data collection agent completed successfully")

		// Step 3: Synthesis Agent
		agentTrace = append(agentTrace, "Starting synthesis agent...")
		result, err := synthesisAgent(ctx, g, input, plan, data, userInput)
		if err != nil {
			return nil, fmt.Errorf("synthesis agent failed: %w", err)
		}
		agentTrace = append(agentTrace, "Synthesis agent completed successfully")

		// Add trace to result
		result.AgentTrace = agentTrace

		return result, nil
	})

	return researchFlow
}

// Streaming Research Flow for real-time progress updates
func StreamingResearchFlow(g *genkit.Genkit) *core.Flow[*ResearchInput, *ResearchResult, string] {
	return StreamingResearchFlowWithUserInput(g, &DefaultUserInputProvider{})
}

// Streaming Research Flow with User Input Provider
func StreamingResearchFlowWithUserInput(g *genkit.Genkit, userInput UserInputProvider) *core.Flow[*ResearchInput, *ResearchResult, string] {
	return genkit.DefineStreamingFlow(g, "streamingResearchFlow",
		func(ctx context.Context, input *ResearchInput, callback core.StreamCallback[string]) (*ResearchResult, error) {
			var agentTrace []string

			// Step 1: Planning Agent
			callback(ctx, "📋 Starting planning agent...")
			agentTrace = append(agentTrace, "Starting planning agent...")

			plan, err := planningAgent(ctx, g, input, userInput)
			if err != nil {
				return nil, fmt.Errorf("planning agent failed: %w", err)
			}

			callback(ctx, "✅ Planning agent completed successfully")
			agentTrace = append(agentTrace, "Planning agent completed successfully")

			// Step 2: Data Collection Agent
			callback(ctx, "📊 Starting data collection agent...")
			agentTrace = append(agentTrace, "Starting data collection agent...")

			data, err := dataCollectionAgent(ctx, g, input, plan, userInput)
			if err != nil {
				return nil, fmt.Errorf("data collection agent failed: %w", err)
			}

			callback(ctx, "✅ Data collection agent completed successfully")
			agentTrace = append(agentTrace, "Data collection agent completed successfully")

			// Step 3: Synthesis Agent
			callback(ctx, "📝 Starting synthesis agent...")
			agentTrace = append(agentTrace, "Starting synthesis agent...")

			result, err := synthesisAgent(ctx, g, input, plan, data, userInput)
			if err != nil {
				return nil, fmt.Errorf("synthesis agent failed: %w", err)
			}

			callback(ctx, "✅ Synthesis agent completed successfully")
			agentTrace = append(agentTrace, "Synthesis agent completed successfully")

			result.AgentTrace = agentTrace

			callback(ctx, "🎉 Research completed successfully!")
			return result, nil
		})
}
