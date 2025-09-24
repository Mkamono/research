package flow

import (
	"context"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"google.golang.org/genai"
)

type DeepResearchInput struct {
	Topic    string `json:"topic" jsonschema:"description=調査したいトピック"`
	Language string `json:"language,omitempty" jsonschema:"description=出力言語,default=日本語"`
}

type ChapterInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Importance  string `json:"importance"`
}

type PlanningResult struct {
	KeyQuestions     []string      `json:"keyQuestions"`
	ResearchApproach string        `json:"researchApproach"`
	Scope            string        `json:"scope"`
	Objectives       string        `json:"objectives"`
	ChapterStructure []ChapterInfo `json:"chapterStructure"`
}

type PlanConfirmationResult struct {
	Approved    bool   `json:"approved"`
	RevisedPlan string `json:"revisedPlan"`
}

type ResearchResult struct {
	Findings       string   `json:"findings"`
	Data           string   `json:"data"`
	ExpertOpinions string   `json:"expertOpinions"`
	SourceUrls     []string `json:"sourceUrls"`
}

type ChapterContent struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	Importance string `json:"importance"`
}

type SynthesisResult struct {
	Chapters         []ChapterContent `json:"chapters"`
	StructureChanges string           `json:"structureChanges"`
}

type SummaryResult struct {
	KeyPoints       []string `json:"keyPoints"`
	Recommendations []string `json:"recommendations"`
}

type DeepResearchResult struct {
	Topic           string   `json:"topic"`
	ResearchPlan    string   `json:"research_plan"`
	KeyQuestions    []string `json:"key_questions"`
	DetailedReport  string   `json:"detailed_report"`
	Sources         []string `json:"sources"`
	Summary         string   `json:"summary"`
	Recommendations string   `json:"recommendations"`
}

// planningPhase performs initial research planning using MCP tools for user interaction
func planningPhase(ctx context.Context, g *genkit.Genkit, input *DeepResearchInput, toolRefs []ai.ToolRef, language string) (*PlanningResult, error) {
	planningPrompt := genkit.LookupPrompt(g, "planning")
	if planningPrompt == nil {
		return nil, fmt.Errorf("planning prompt not found")
	}

	resp, err := planningPrompt.Execute(ctx,
		ai.WithInput(map[string]any{
			"topic":    input.Topic,
			"language": language,
		}),
		ai.WithTools(toolRefs...))
	if err != nil {
		return nil, fmt.Errorf("planning phase failed: %w", err)
	}

	var result PlanningResult
	if err := resp.Output(&result); err != nil {
		return nil, fmt.Errorf("failed to parse planning result: %w", err)
	}

	return &result, nil
}

// planConfirmationPhase asks user to confirm the research plan using ask-me tool and iterates until approval
func planConfirmationPhase(ctx context.Context, g *genkit.Genkit, initialPlan string, toolRefs []ai.ToolRef, language string) (string, error) {
	confirmationPrompt := genkit.LookupPrompt(g, "plan_confirmation")
	if confirmationPrompt == nil {
		return "", fmt.Errorf("plan_confirmation prompt not found")
	}

	currentPlan := initialPlan
	maxIterations := 10

	for i := 0; i < maxIterations; i++ {
		resp, err := confirmationPrompt.Execute(ctx,
			ai.WithInput(map[string]any{
				"currentPlan": currentPlan,
				"isFirstTime": i == 0,
				"language":    language,
			}),
			ai.WithTools(toolRefs...))
		if err != nil {
			// If API error, try with simpler prompt without tools
			if strings.Contains(err.Error(), "INTERNAL") {
				simplePrompt := fmt.Sprintf("調査計画を確認してください: %s", currentPlan)
				resp, err = genkit.Generate(ctx, g,
					ai.WithConfig(&genai.GenerateContentConfig{
						Temperature: genai.Ptr[float32](0.3),
					}),
					ai.WithPrompt(simplePrompt))
				if err != nil {
					return "", fmt.Errorf("plan confirmation phase failed at iteration %d (even with simple prompt): %w", i+1, err)
				}
				// For simple prompt fallback, use text response
				return resp.Text(), nil
			} else {
				return "", fmt.Errorf("plan confirmation phase failed at iteration %d: %w", i+1, err)
			}
		}

		var result PlanConfirmationResult
		parseError := resp.Output(&result)

		if parseError != nil {
			// Fallback to text parsing if structured output fails
			responseText := resp.Text()
			fmt.Printf("Failed to parse JSON response, using text fallback. Error: %v\nResponse: %s\n", parseError, responseText)

			// Check for approval indicators in the text
			if strings.Contains(responseText, "承認") || strings.Contains(responseText, "APPROVED") ||
				strings.Contains(responseText, "approved") || strings.Contains(responseText, "OK") ||
				strings.Contains(strings.ToLower(responseText), "approved") {
				return currentPlan, nil
			}

			// If no clear approval, treat the response as a revised plan
			if responseText != "" {
				currentPlan = responseText
			}
			continue
		}

		// Successfully parsed JSON response
		if result.Approved {
			return currentPlan, nil
		} else if !result.Approved && result.RevisedPlan != "" {
			currentPlan = result.RevisedPlan
			continue
		}

		// If we reach here, try to use the text response as fallback
		responseText := resp.Text()
		if responseText != "" {
			currentPlan = responseText
		}
	}

	return currentPlan, fmt.Errorf("maximum iterations (%d) reached for plan confirmation, proceeding with last plan", maxIterations)
}

// researchPhase performs detailed web search for each research question
func researchPhase(ctx context.Context, g *genkit.Genkit, keyQuestions []string, language string) ([]string, []string, error) {
	researchPrompt := genkit.LookupPrompt(g, "research")
	if researchPrompt == nil {
		return nil, nil, fmt.Errorf("research prompt not found")
	}

	var allFindings []string
	var sources []string

	for _, question := range keyQuestions {
		resp, err := researchPrompt.Execute(ctx,
			ai.WithInput(map[string]any{
				"question": question,
				"language": language,
			}))
		if err != nil {
			return nil, nil, fmt.Errorf("web search failed for question '%s': %w", question, err)
		}

		var result ResearchResult
		if err := resp.Output(&result); err != nil {
			// Fallback to text if structured output fails
			allFindings = append(allFindings, fmt.Sprintf("【%s】\n%s", question, resp.Text()))
			sources = append(sources, fmt.Sprintf("Search results for: %s", question))
			continue
		}

		// Format the structured result
		formattedResult := fmt.Sprintf("【%s】\n主要な発見事項: %s\n重要なデータ: %s\n専門家の意見: %s",
			question, result.Findings, result.Data, result.ExpertOpinions)
		allFindings = append(allFindings, formattedResult)

		// Add source URLs
		for _, url := range result.SourceUrls {
			sources = append(sources, url)
		}
	}

	return allFindings, sources, nil
}

// synthesisPhase creates the final comprehensive report and summary
func synthesisPhase(ctx context.Context, g *genkit.Genkit, input *DeepResearchInput, researchPlan string, allFindings []string, chapterStructure []ChapterInfo, language string) (string, string, error) {
	// Generate comprehensive report
	synthesisPrompt := genkit.LookupPrompt(g, "synthesis")
	if synthesisPrompt == nil {
		return "", "", fmt.Errorf("synthesis prompt not found")
	}

	// Format chapter structure for prompt
	chapterStructureText := ""
	for i, chapter := range chapterStructure {
		chapterStructureText += fmt.Sprintf("%d. %s (%s重要度)\n   %s\n\n",
			i+1, chapter.Title, chapter.Importance, chapter.Description)
	}

	synthesisResp, err := synthesisPrompt.Execute(ctx,
		ai.WithInput(map[string]any{
			"topic":             input.Topic,
			"investigationPlan": researchPlan,
			"chapterStructure":  chapterStructureText,
			"allFindings":       strings.Join(allFindings, "\n\n"),
			"language":          language,
		}))
	if err != nil {
		return "", "", fmt.Errorf("synthesis failed: %w", err)
	}

	var synthesisResult SynthesisResult
	var detailedReport string
	if err := synthesisResp.Output(&synthesisResult); err != nil {
		// Fallback to text if structured output fails
		detailedReport = synthesisResp.Text()
	} else {
		// Format the structured result into a report
		var reportBuilder strings.Builder
		for i, chapter := range synthesisResult.Chapters {
			reportBuilder.WriteString(fmt.Sprintf("%d. %s\n%s\n\n", i+1, chapter.Title, chapter.Content))
		}
		detailedReport = reportBuilder.String()

		// Add structure changes note if any
		if synthesisResult.StructureChanges != "" {
			detailedReport += fmt.Sprintf("\n--- 章構成の変更点 ---\n%s\n", synthesisResult.StructureChanges)
		}
	}

	// Generate summary and recommendations
	summaryPrompt := genkit.LookupPrompt(g, "summary")
	if summaryPrompt == nil {
		return "", "", fmt.Errorf("summary prompt not found")
	}

	summaryResp, err := summaryPrompt.Execute(ctx,
		ai.WithInput(map[string]any{
			"detailedReport": detailedReport,
			"language":       language,
		}))
	if err != nil {
		return "", "", fmt.Errorf("summary generation failed: %w", err)
	}

	var summaryResult SummaryResult
	var summaryText string
	if err := summaryResp.Output(&summaryResult); err != nil {
		// Fallback to text if structured output fails
		summaryText = summaryResp.Text()
	} else {
		// Format the structured result
		summaryText = fmt.Sprintf("重要なポイント:\n%s\n\n推奨事項:\n%s",
			strings.Join(summaryResult.KeyPoints, "\n"),
			strings.Join(summaryResult.Recommendations, "\n"))
	}

	return detailedReport, summaryText, nil
}

// reportDeliveryPhase delivers the final research report to user using ask-me tool
func reportDeliveryPhase(ctx context.Context, g *genkit.Genkit, result *DeepResearchResult, toolRefs []ai.ToolRef, language string) error {
	reportDeliveryPrompt := genkit.LookupPrompt(g, "report_delivery")
	if reportDeliveryPrompt == nil {
		return fmt.Errorf("report_delivery prompt not found")
	}

	_, err := reportDeliveryPrompt.Execute(ctx,
		ai.WithInput(map[string]any{
			"topic":           result.Topic,
			"detailedReport":  result.DetailedReport,
			"summary":         result.Summary,
			"recommendations": result.Recommendations,
			"language":        language,
		}),
		ai.WithTools(toolRefs...))
	if err != nil {
		return fmt.Errorf("report delivery phase failed: %w", err)
	}

	return nil
}

func DeepResearchFlow(g *genkit.Genkit, mcpTools []ai.Tool) *core.Flow[*DeepResearchInput, *DeepResearchResult, struct{}] {
	return genkit.DefineFlow(g, "deepResearchFlow", func(ctx context.Context, input *DeepResearchInput) (*DeepResearchResult, error) {
		language := input.Language
		if language == "" {
			language = "日本語" // Default to Japanese
		}

		// Convert MCP tools to ToolRef
		toolRefs := make([]ai.ToolRef, len(mcpTools))
		for i, tool := range mcpTools {
			toolRefs[i] = tool
		}

		// Phase 1: Research planning with user interaction
		planningResult, err := planningPhase(ctx, g, input, toolRefs, language)
		if err != nil {
			return nil, err
		}

		// Create initial plan text from structured result
		initialPlan := fmt.Sprintf("調査の目的: %s\n調査の範囲: %s\n調査アプローチ: %s\n重要な質問: %s",
			planningResult.Objectives,
			planningResult.Scope,
			planningResult.ResearchApproach,
			strings.Join(planningResult.KeyQuestions, ", "))

		// Phase 2: Plan confirmation with user
		researchPlan, err := planConfirmationPhase(ctx, g, initialPlan, toolRefs, language)
		if err != nil {
			return nil, err
		}

		// Phase 3: Use key research questions from planning phase
		keyQuestions := planningResult.KeyQuestions
		if len(keyQuestions) == 0 {
			// Fallback to default questions if none provided
			keyQuestions = []string{
				fmt.Sprintf("%sに関する最新の調査", input.Topic),
				fmt.Sprintf("%sの現在の課題と問題点", input.Topic),
				fmt.Sprintf("%sの将来的な展望", input.Topic),
			}
		}

		// Phase 4: Detailed web search research
		allFindings, sources, err := researchPhase(ctx, g, keyQuestions, language)
		if err != nil {
			return nil, err
		}

		// Phase 5: Synthesis and final report generation
		detailedReport, summary, err := synthesisPhase(ctx, g, input, researchPlan, allFindings, planningResult.ChapterStructure, language)
		if err != nil {
			return nil, err
		}

		// Create the result object
		result := &DeepResearchResult{
			Topic:           input.Topic,
			ResearchPlan:    researchPlan,
			KeyQuestions:    keyQuestions,
			DetailedReport:  detailedReport,
			Sources:         sources,
			Summary:         summary,
			Recommendations: summary, // In practice, you'd parse this separately
		}

		// Phase 6: Report delivery to user using ask-me tool
		err = reportDeliveryPhase(ctx, g, result, toolRefs, language)
		if err != nil {
			return nil, err
		}

		return result, nil
	})
}
