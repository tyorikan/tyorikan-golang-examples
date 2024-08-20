package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/vertexai"
)

type MenuSuggestion struct {
	RestaurantName    string `json:"restaurant_name"`
	RestaurantConcept string `json:"restaurant_concept"`
	Menus             []Menu `json:"menus"`
}

type Menu struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	Category    string `json:"category"`
}

type Category int

const (
	Appetizer Category = iota
	MainDish
	Rice
	Dessert
	Drink
)

func (c Category) String() string {
	switch c {
	case Appetizer:
		return "前菜・一品料理"
	case MainDish:
		return "メイン料理"
	case Rice:
		return "ご飯もの・麺類"
	case Dessert:
		return "デザート"
	case Drink:
		return "ドリンク"
	default:
		return "その他"
	}
}

func main() {
	ctx := context.Background()

	// Initialize Vertex AI.
	if err := vertexai.Init(ctx, &vertexai.Config{Location: "asia-northeast1"}); err != nil {
		log.Fatal(err)
	}

	// Define a Genkit flow for menu suggestion generation.
	genkit.DefineFlow("menuSuggestionFlow", getMenuSuggestion)

	// Initialize Genkit.
	if err := genkit.Init(ctx, nil); err != nil {
		log.Fatal(err)
	}
}

// getMenuSuggestion generates a menu suggestion based on the provided restaurant theme.
func getMenuSuggestion(ctx context.Context, restaurantTheme string) (*MenuSuggestion, error) {
	m := vertexai.Model("gemini-1.5-flash")

	// Construct the prompt for the AI model.
	prompt := fmt.Sprintf(`
		%sをテーマにしたレストランのオリジナル メニューを提案して。
		category は「%s」のどれかになるようにして。
	`,
		// %sをテーマにしたレストランのメニューを// 少なくとも 20 品以上提案して。
		// 1 品あたりの金額は、400 円から 3,000 円までの範囲にして。
		restaurantTheme,
		strings.Join([]string{Appetizer.String(), MainDish.String(), Rice.String(), Dessert.String(), Drink.String()}, ","),
	)

	// Convert the sample menu suggestion to a map for use as the schema.
	menuSuggestionMap, err := generateMenuSchema()
	if err != nil {
		return nil, err
	}

	// Create the GenerateRequest with the prompt and schema.
	req := ai.NewGenerateRequest(
		&ai.GenerationCommonConfig{Temperature: 1},
		ai.NewUserTextMessage(prompt),
	)
	req.Output = &ai.GenerateRequestOutput{
		Format: ai.OutputFormatJSON,
		Schema: *menuSuggestionMap,
	}

	// Generate the menu suggestion using the AI model.
	resp, err := m.Generate(ctx, req, nil)
	if err != nil {
		return nil, err
	}

	// Unmarshal the AI response into a MenuSuggestion struct.
	menu := &MenuSuggestion{}
	err = resp.UnmarshalOutput(menu)
	return menu, err
}

func generateMenuSchema() (*map[string]any, error) {
	// Create a sample menu suggestion to use as a schema for the AI response.
	sampleMenuSuggestion := &MenuSuggestion{
		RestaurantName:    "イタリアンレストラン",
		RestaurantConcept: "本格イタリアン",
		Menus: []Menu{
			{
				Name:        "パスタ",
				Description: "トマトソースのパスタ",
				Price:       1000,
				Category:    Rice.String(),
			},
		},
	}

	// Convert the sample menu suggestion to a map for use as the schema.
	menuSuggestionMap := make(map[string]any)
	data, err := json.Marshal(sampleMenuSuggestion)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &menuSuggestionMap)
	if err != nil {
		return nil, err
	}

	return &menuSuggestionMap, nil
}
