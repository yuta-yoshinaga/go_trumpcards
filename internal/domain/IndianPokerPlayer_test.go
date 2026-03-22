//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIndianPokerPlayer(t *testing.T) {
	t.Run("human player", func(t *testing.T) {
		p := NewIndianPokerPlayer(true, HoldemStyleTAG)
		assert.True(t, p.GetIsHuman())
		assert.Equal(t, HoldemStyleTAG, p.GetPlayStyle())
	})

	t.Run("CPU player", func(t *testing.T) {
		p := NewIndianPokerPlayer(false, HoldemStyleLAP)
		assert.False(t, p.GetIsHuman())
		assert.Equal(t, HoldemStyleLAP, p.GetPlayStyle())
	})
}

func TestIndianPokerPlayer_GetPlayStyleName(t *testing.T) {
	tests := []struct {
		style HoldemPlayStyle
		name  string
	}{
		{HoldemStyleTAG, "TAG"},
		{HoldemStyleLAP, "LAP"},
		{HoldemStyleTAP, "TAP"},
		{HoldemStyleLAG, "LAG"},
		{HoldemStyleGTO, "GTO"},
		{HoldemPlayStyle(99), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewIndianPokerPlayer(false, tt.style)
			assert.Equal(t, tt.name, p.GetPlayStyleName())
		})
	}
}

func TestIndianPokerPlayer_GetComparisonCards(t *testing.T) {
	t.Run("no cards returns nil", func(t *testing.T) {
		p := NewIndianPokerPlayer(true, HoldemStyleTAG)
		assert.Nil(t, p.GetComparisonCards())
	})

	t.Run("one card returns single card slice", func(t *testing.T) {
		p := NewIndianPokerPlayer(true, HoldemStyleTAG)
		card := NewCard(CardDesignSpade, 10, false)
		p.AddCard(card)
		result := p.GetComparisonCards()
		assert.Len(t, result, 1)
		assert.Equal(t, card, result[0])
	})
}

func TestIndianPokerPlayer_ChipHolder(t *testing.T) {
	p := NewIndianPokerPlayer(true, HoldemStyleTAG)

	// Initial chips = 0
	assert.Equal(t, 0, p.GetChips())

	// SetChips
	p.SetChips(1000)
	assert.Equal(t, 1000, p.GetChips())

	// AddChips
	p.AddChips(500)
	assert.Equal(t, 1500, p.GetChips())

	// SubtractChips success
	ok := p.SubtractChips(200)
	assert.True(t, ok)
	assert.Equal(t, 1300, p.GetChips())

	// SubtractChips failure (insufficient)
	ok = p.SubtractChips(2000)
	assert.False(t, ok)
	assert.Equal(t, 1300, p.GetChips())
}

func TestIndianPokerPlayer_BettingState(t *testing.T) {
	p := NewIndianPokerPlayer(false, HoldemStyleTAG)

	// Folded
	assert.False(t, p.GetFolded())
	p.SetFolded(true)
	assert.True(t, p.GetFolded())

	// AllIn
	assert.False(t, p.GetAllIn())
	p.SetAllIn(true)
	assert.True(t, p.GetAllIn())

	// CurrentBet
	assert.Equal(t, 0, p.GetCurrentBet())
	p.SetCurrentBet(50)
	assert.Equal(t, 50, p.GetCurrentBet())

	// HandRank
	assert.Equal(t, 0, p.GetHandRank())
	p.SetHandRank(14)
	assert.Equal(t, 14, p.GetHandRank())
}
