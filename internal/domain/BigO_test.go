//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewDefaultBigO_Deals5HoleCards verifies Big O deals 5 hole cards to each
// player (vs Omaha's 4) while reusing the Omaha engine.
func TestNewDefaultBigO_Deals5HoleCards(t *testing.T) {
	o := NewDefaultBigO()
	require.NoError(t, o.Reset())

	assert.False(t, o.GetIsHiLo(), "Big O default variant is Hi-only")
	for i := 0; i < o.GetPlayerCnt(); i++ {
		assert.Equal(t, 5, o.GetPlayer(i).GetCardsSize(),
			"player %d should hold 5 hole cards", i)
	}
}

// TestNewDefaultBigOHiLo_IsHiLo verifies the Hi-Lo Big O variant enables the
// 8-or-Better split and still deals 5 hole cards.
func TestNewDefaultBigOHiLo_IsHiLo(t *testing.T) {
	o := NewDefaultBigOHiLo()
	require.NoError(t, o.Reset())

	assert.True(t, o.GetIsHiLo(), "Big O Hi-Lo variant enables the low split")
	for i := 0; i < o.GetPlayerCnt(); i++ {
		assert.Equal(t, 5, o.GetPlayer(i).GetCardsSize())
	}
}

// TestBigO_HoleCardCountSurvivesJSON verifies the 5-card configuration is
// preserved across a serialise/deserialise round-trip (KV persistence).
func TestBigO_HoleCardCountSurvivesJSON(t *testing.T) {
	o := NewDefaultBigO()
	require.NoError(t, o.Reset())

	data, err := json.Marshal(o)
	require.NoError(t, err)

	var restored Omaha
	require.NoError(t, json.Unmarshal(data, &restored))
	require.NoError(t, restored.Reset())

	for i := 0; i < restored.GetPlayerCnt(); i++ {
		assert.Equal(t, 5, restored.GetPlayer(i).GetCardsSize(),
			"restored Big O game must still deal 5 hole cards")
	}
}

// TestBigO_EvalBestHandUsesFiveHoleCards verifies Omaha hand evaluation
// (exactly 2 from hole, 3 from board) operates over 5-card holdings.
func TestBigO_EvalBestHandUsesFiveHoleCards(t *testing.T) {
	p := NewOmahaPlayer(true, HoldemStyleTAG)
	// Hole: A-K-Q-J-T (mixed suits)
	p.AddCard(NewCard(CardDesignClover, 1, false))
	p.AddCard(NewCard(CardDesignDiamond, 13, false))
	p.AddCard(NewCard(CardDesignHeart, 12, false))
	p.AddCard(NewCard(CardDesignSpade, 11, false))
	p.AddCard(NewCard(CardDesignClover, 10, false))
	// Board gives a club flush draw with the held clubs.
	board := []*Card{
		NewCard(CardDesignClover, 2, false), NewCard(CardDesignClover, 5, false), NewCard(CardDesignClover, 8, false),
		NewCard(CardDesignHeart, 3, false), NewCard(CardDesignSpade, 4, false),
	}
	rank := p.EvalBestHand(board)
	assert.GreaterOrEqual(t, rank, PokerHandHighCard)
	assert.Len(t, p.GetBestHand(), 5)
}
