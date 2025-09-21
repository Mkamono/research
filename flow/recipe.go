package flow

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

// Define input schema
type RecipeInput struct {
	Ingredient          string `json:"ingredient" jsonschema:"description=Main ingredient or cuisine type"`
	DietaryRestrictions string `json:"dietaryRestrictions,omitempty" jsonschema:"description=Any dietary restrictions"`
}

// Define output schema
type Recipe struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	PrepTime     string   `json:"prepTime"`
	CookTime     string   `json:"cookTime"`
	Servings     int      `json:"servings"`
	Ingredients  []string `json:"ingredients"`
	Instructions []string `json:"instructions"`
	Tips         []string `json:"tips,omitempty"`
}

func RecipeGeneratorFlow(g *genkit.Genkit) *core.Flow[*RecipeInput, *Recipe, struct{}] {
	// Define a recipe generator flow
	recipeGeneratorFlow := genkit.DefineFlow(g, "recipeGeneratorFlow", func(ctx context.Context, input *RecipeInput) (*Recipe, error) {
		// Create a prompt based on the input
		dietaryRestrictions := input.DietaryRestrictions
		if dietaryRestrictions == "" {
			dietaryRestrictions = "none"
		}

		prompt := fmt.Sprintf(`Create a recipe with the following requirements:
            Main ingredient: %s
            Dietary restrictions: %s`, input.Ingredient, dietaryRestrictions)

		// Generate structured recipe data using the same schema
		recipe, _, err := genkit.GenerateData[Recipe](ctx, g,
			ai.WithPrompt(prompt),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to generate recipe: %w", err)
		}

		return recipe, nil
	})

	return recipeGeneratorFlow
}
