//go:build test

package domain_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

type shengJiComboGoldenCard struct {
	Design string `json:"design"`
	Value  int    `json:"value"`
}

type shengJiComboGoldenCase struct {
	Name      string                   `json:"name"`
	Cards     []shengJiComboGoldenCard `json:"cards"`
	Level     int                      `json:"level"`
	TrumpSuit int                      `json:"trumpSuit"`
	Kind      int                      `json:"kind"`
	Rank      int                      `json:"rank"`
	Size      int                      `json:"size"`
	Trump     bool                     `json:"trump"`
}

func TestShengJiComboGoldenValues(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "frontend", "src", "constants", "shengjiCombo.json"))
	require.NoError(t, err)

	var golden []shengJiComboGoldenCase
	require.NoError(t, json.Unmarshal(raw, &golden))
	require.GreaterOrEqual(t, len(golden), 30)
	require.LessOrEqual(t, len(golden), 60)

	designs := map[string]int{
		"JOKER":   domain.CardDesignJoker,
		"SPADE":   domain.CardDesignSpade,
		"CLOVER":  domain.CardDesignClover,
		"HEART":   domain.CardDesignHeart,
		"DIAMOND": domain.CardDesignDiamond,
	}
	for _, tc := range golden {
		t.Run(tc.Name, func(t *testing.T) {
			cards := make([]*domain.Card, 0, len(tc.Cards))
			for _, goldenCard := range tc.Cards {
				design, ok := designs[goldenCard.Design]
				require.True(t, ok, "unknown design %q", goldenCard.Design)
				cards = append(cards, domain.NewCard(design, goldenCard.Value, true))
			}

			got := domain.ShengJiEvaluate(cards, tc.Level, tc.TrumpSuit)
			if tc.Kind == int(domain.ShengJiComboNone) {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, tc.Kind, int(got.Kind))
			require.Equal(t, tc.Rank, got.Rank)
			require.Equal(t, tc.Size, got.Size)
			// TS 側 (shengjiCombo.test.ts) も trump を見る。**両側が同じだけ検査する**のがこのガードの前提。
			require.Equal(t, tc.Trump, got.Trump)
		})
	}
}
