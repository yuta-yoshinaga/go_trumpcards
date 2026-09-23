//go:build test

package domain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type spideretteGoldenCard struct {
	Suit   int  `json:"suit"`
	Value  int  `json:"value"`
	FaceUp bool `json:"faceUp"`
}

type spideretteGoldenCase struct {
	Name        string                 `json:"name"`
	Cards       []spideretteGoldenCard `json:"cards"`
	SourceIndex int                    `json:"sourceIndex"`
	Valid       bool                   `json:"valid"`
}

type spideretteGolden struct {
	Cases []spideretteGoldenCase `json:"cases"`
}

func TestSpideretteValidSequence_GoldenVectors(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(
		"..", "..", "frontend", "src", "utils", "__fixtures__", "spideretteMoves.golden.json"))
	require.NoError(t, err)
	var golden spideretteGolden
	require.NoError(t, json.Unmarshal(raw, &golden))
	require.NotEmpty(t, golden.Cases)

	designs := []int{CardDesignJoker, CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	for _, c := range golden.Cases {
		t.Run(c.Name, func(t *testing.T) {
			cards := make([]*SpideretteTableauCard, 0, len(c.Cards))
			for _, card := range c.Cards {
				cards = append(cards, &SpideretteTableauCard{
					Card:   NewCard(designs[card.Suit], card.Value, false),
					FaceUp: card.FaceUp,
				})
			}
			got := cards[c.SourceIndex].FaceUp && (&Spiderette{}).isValidSequence(cards[c.SourceIndex:])
			require.Equal(t, c.Valid, got)
		})
	}
}
