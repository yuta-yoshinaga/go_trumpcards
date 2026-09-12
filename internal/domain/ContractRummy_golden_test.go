//go:build test

package domain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type contractRummyGoldenCard struct {
	Suit  int `json:"suit"`
	Value int `json:"value"`
}

type contractRummyGoldenCase struct {
	Name  string                    `json:"name"`
	Meld  []contractRummyGoldenCard `json:"meld"`
	Card  contractRummyGoldenCard   `json:"card"`
	Valid bool                      `json:"valid"`
}

type contractRummyGolden struct {
	Cases []contractRummyGoldenCase `json:"cases"`
}

func TestContractRummyCanAddToMeld_GoldenVectors(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "frontend", "src", "utils", "__fixtures__", "contractRummyLayoff.golden.json"))
	require.NoError(t, err)
	var golden contractRummyGolden
	require.NoError(t, json.Unmarshal(raw, &golden))
	require.NotEmpty(t, golden.Cases)

	designs := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	for _, c := range golden.Cases {
		t.Run(c.Name, func(t *testing.T) {
			meld := make([]*Card, 0, len(c.Meld))
			for _, card := range c.Meld {
				meld = append(meld, NewCard(designs[card.Suit], card.Value, false))
			}
			card := NewCard(designs[c.Card.Suit], c.Card.Value, false)
			require.Equal(t, c.Valid, canAddToContractRummyMeld(meld, card))
		})
	}
}
