//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Baker's Game (ベーカーズ・ゲーム) reuses the FreeCell engine with one rule
// change: tableau stacking requires the SAME suit in descending order instead of
// alternating colour. These tests pin that single difference plus the flag's
// persistence, leaving the shared FreeCell behaviour to FreeCell_test.go.

func setupPlayingBakersGame() *FreeCell {
	b := NewDefaultBakersGame()
	b.Reset()
	return b
}

func TestNewDefaultBakersGameUsesSameSuit(t *testing.T) {
	b := NewDefaultBakersGame()
	assert.True(t, b.sameSuit, "Baker's Game must stack by same suit")
	// FreeCell remains alternating colour.
	assert.False(t, NewDefaultFreeCell().sameSuit)
}

func TestBakersGameCanPlaceOnTableau(t *testing.T) {
	b := setupPlayingBakersGame()
	clearTableauFC(b)

	t.Run("any card on empty column", func(t *testing.T) {
		assert.True(t, b.canPlaceOnTableau(makeCard(CardDesignSpade, 5), 0))
	})

	t.Run("same suit descending allowed", func(t *testing.T) {
		b.tableau[0] = []*Card{makeCard(CardDesignSpade, 8)}
		assert.True(t, b.canPlaceOnTableau(makeCard(CardDesignSpade, 7), 0))
	})

	t.Run("alternating colour rejected", func(t *testing.T) {
		b.tableau[1] = []*Card{makeCard(CardDesignSpade, 8)}
		// In FreeCell this would be legal; in Baker's Game it is not.
		assert.False(t, b.canPlaceOnTableau(makeCard(CardDesignHeart, 7), 1))
	})

	t.Run("same suit but not descending rejected", func(t *testing.T) {
		b.tableau[2] = []*Card{makeCard(CardDesignSpade, 8)}
		assert.False(t, b.canPlaceOnTableau(makeCard(CardDesignSpade, 6), 2))
	})
}

func TestBakersGameMoveTableauSameSuitOnly(t *testing.T) {
	b := setupPlayingBakersGame()
	clearTableauFC(b)
	b.tableau[0] = []*Card{makeCard(CardDesignSpade, 8)}
	b.tableau[1] = []*Card{makeCard(CardDesignHeart, 7)}
	b.tableau[2] = []*Card{makeCard(CardDesignSpade, 7)}

	// Heart 7 cannot move onto Spade 8 (different suit).
	assert.Error(t, b.MoveTableauToTableau(1, 0, 0))
	// Spade 7 can move onto Spade 8 (same suit, descending).
	require.NoError(t, b.MoveTableauToTableau(2, 0, 0))
	assert.Len(t, b.tableau[0], 2)
	assert.Empty(t, b.tableau[2])
}

func TestBakersGameIsValidTableauSequence(t *testing.T) {
	b := NewDefaultBakersGame()
	sameSuit := []*Card{makeCard(CardDesignClover, 9), makeCard(CardDesignClover, 8), makeCard(CardDesignClover, 7)}
	altColour := []*Card{makeCard(CardDesignClover, 9), makeCard(CardDesignHeart, 8)}
	assert.True(t, b.isValidTableauSequence(sameSuit))
	assert.False(t, b.isValidTableauSequence(altColour))
}

func TestBakersGamePersistsSameSuitFlag(t *testing.T) {
	original := setupPlayingBakersGame()
	require.NoError(t, original.MoveTableauToFreeCell(0, 0))

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored FreeCell
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.True(t, restored.sameSuit, "sameSuit flag must round-trip through JSON")
}
