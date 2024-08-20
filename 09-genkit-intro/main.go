package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

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

	cfg := vertexai.Config{
		Location: "asia-northeast1",
	}
	if err := vertexai.Init(ctx, &cfg); err != nil {
		log.Fatal(err)
	}

	genkit.DefineFlow("menuSuggestionFlow", func(ctx context.Context, restaurantTheme string) (*MenuSuggestion, error) {
		m := vertexai.Model("gemini-1.5-flash")
		if m == nil {
			return nil, errors.New("menuSuggestionFlow: failed to find model")
		}

		req := ai.NewGenerateRequest(
			&ai.GenerationCommonConfig{Temperature: 1},
			ai.NewUserTextMessage(fmt.Sprintf(`%sをテーマにしたレストランのメニューを提案して`, restaurantTheme)),
		)

		menus := []Menu{
			{Name: "パスタ", Description: "トマトソースのパスタ", Price: 1000, Category: Rice.String()},
			{Name: "ピザ", Description: "マルゲリータピザ", Price: 1200, Category: Rice.String()},
		}

		menuSuggestion := &MenuSuggestion{
			RestaurantName:    "イタリアンレストラン",
			RestaurantConcept: "本格イタリアン",
			Menus:             menus,
		}

		menuSuggestionMap := make(map[string]any)
		data, err := json.Marshal(menuSuggestion)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(data, &menuSuggestionMap)
		if err != nil {
			return nil, err
		}

		req.Output = &ai.GenerateRequestOutput{
			Format: ai.OutputFormatJSON,
			Schema: menuSuggestionMap,
		}

		resp, err := m.Generate(ctx, req, nil)
		if err != nil {
			return nil, err
		}

		menu := &MenuSuggestion{}
		err = resp.UnmarshalOutput(menu)
		return menu, err
	})

	if err := genkit.Init(ctx, nil); err != nil {
		log.Fatal(err)
	}
}
