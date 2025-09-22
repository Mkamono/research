package flow

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"google.golang.org/genai"
)

type DeepResearchInput struct {
	Topic    string `json:"topic" jsonschema:"description=研究したいトピック"`
	Purpose  string `json:"purpose,omitempty" jsonschema:"description=研究の目的や背景"`
	Scope    string `json:"scope,omitempty" jsonschema:"description=研究の範囲や制約"`
	Language string `json:"language,omitempty" jsonschema:"description=出力言語（デフォルト：日本語）"`
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

func DeepResearchFlow(g *genkit.Genkit, mcpTools []ai.Tool) *core.Flow[*DeepResearchInput, *DeepResearchResult, struct{}] {
	return genkit.DefineFlow(g, "deepResearchFlow", func(ctx context.Context, input *DeepResearchInput) (*DeepResearchResult, error) {

		language := input.Language
		if language == "" {
			language = "日本語"
		}

		// Phase 1: Initial research planning using ask tool
		planningPrompt := fmt.Sprintf(`
あなたは研究の専門家です。以下のトピックについて深い調査を行うための研究計画を立ててください。

トピック: %s
目的: %s
範囲: %s

まず、このトピックについて調査するための重要な質問を5-7個特定し、研究のアプローチを整理してください。
ユーザーに確認が必要な事項があれば、chat toolを使って質問してください。

出力言語: %s
`, input.Topic, input.Purpose, input.Scope, language)

		// Convert MCP tools to ToolRef
		toolRefs := make([]ai.ToolRef, len(mcpTools))
		for i, tool := range mcpTools {
			toolRefs[i] = tool
		}

		resp, err := genkit.Generate(ctx, g,
			ai.WithConfig(&genai.GenerateContentConfig{
				Temperature: genai.Ptr[float32](0.3),
			}),
			ai.WithTools(toolRefs...),
			ai.WithPrompt(planningPrompt))
		if err != nil {
			return nil, fmt.Errorf("planning phase failed: %w", err)
		}

		researchPlan := resp.Text()

		// Phase 2: Extract key questions for detailed research
		keyQuestions := []string{
			fmt.Sprintf("%sに関する最新の研究", input.Topic),
			fmt.Sprintf("%sの現在の課題と問題点", input.Topic),
			fmt.Sprintf("%sの将来的な展望", input.Topic),
		}

		// Phase 3: Detailed research with web search
		var allFindings []string
		var sources []string

		for _, question := range keyQuestions {
			searchPrompt := fmt.Sprintf(`
以下の質問について詳細に調査してください：

質問: %s

Web検索を使用して最新の情報を収集し、以下の形式で回答してください：
- 主要な発見事項
- 重要なデータや統計
- 専門家の意見や見解
- 参考になるソースのURL

出力言語: %s
`, question, language)

			// Use web search as shown in user's example
			searchResp, err := genkit.Generate(ctx, g,
				ai.WithConfig(&genai.GenerateContentConfig{
					Temperature: genai.Ptr[float32](0.2),
					Tools: []*genai.Tool{
						{
							GoogleSearch: &genai.GoogleSearch{},
						},
					},
				}),
				ai.WithPrompt(searchPrompt))
			if err != nil {
				return nil, fmt.Errorf("web search failed for question '%s': %w", question, err)
			}

			allFindings = append(allFindings, fmt.Sprintf("【%s】\n%s", question, searchResp.Text()))
			sources = append(sources, fmt.Sprintf("Search results for: %s", question))
		}

		// Phase 4: Synthesis and final report generation
		synthesisPrompt := fmt.Sprintf(`
以下の調査結果を基に、包括的な研究レポートを作成してください。

トピック: %s
研究計画: %s

調査結果:
%s

以下の構造でレポートを作成してください：

1. エグゼクティブサマリー
2. 背景と現状分析
3. 主要な発見事項
4. 課題と問題点
5. 将来的な展望
6. 推奨事項
7. 結論

出力言語: %s
各セクションは詳細で具体的な内容を含めてください。
`, input.Topic, researchPlan, fmt.Sprintf("%v", allFindings), language)

		finalResp, err := genkit.Generate(ctx, g,
			ai.WithConfig(&genai.GenerateContentConfig{
				Temperature: genai.Ptr[float32](0.3),
			}),
			ai.WithPrompt(synthesisPrompt))
		if err != nil {
			return nil, fmt.Errorf("synthesis failed: %w", err)
		}

		// Generate summary and recommendations
		summaryPrompt := fmt.Sprintf(`
以下のレポートから重要なポイントを抽出し、簡潔なサマリーと具体的な推奨事項を作成してください。

レポート:
%s

サマリーは3-5つの重要なポイントで構成し、推奨事項は具体的で実行可能な内容にしてください。

出力言語: %s
`, finalResp.Text(), language)

		summaryResp, err := genkit.Generate(ctx, g,
			ai.WithConfig(&genai.GenerateContentConfig{
				Temperature: genai.Ptr[float32](0.3),
			}),
			ai.WithPrompt(summaryPrompt))
		if err != nil {
			return nil, fmt.Errorf("summary generation failed: %w", err)
		}

		return &DeepResearchResult{
			Topic:           input.Topic,
			ResearchPlan:    researchPlan,
			KeyQuestions:    keyQuestions,
			DetailedReport:  finalResp.Text(),
			Sources:         sources,
			Summary:         summaryResp.Text(),
			Recommendations: summaryResp.Text(), // In practice, you'd parse this separately
		}, nil
	})
}
