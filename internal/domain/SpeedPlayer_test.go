//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewSpeedPlayer(t *testing.T) {
	t.Run("human player", func(t *testing.T) {
		p := domain.NewSpeedPlayer(true)
		assert.True(t, p.GetIsHuman())
		assert.Equal(t, 0, p.GetCardsSize())
		assert.Equal(t, 0, p.GetDrawPileSize())
	})
	t.Run("CPU player", func(t *testing.T) {
		p := domain.NewSpeedPlayer(false)
		assert.False(t, p.GetIsHuman())
	})
}

func TestSpeedPlayer_AddToDrawPile(t *testing.T) {
	p := domain.NewSpeedPlayer(true)
	c1 := domain.NewCard(domain.CardDesignSpade, 1, false)
	c2 := domain.NewCard(domain.CardDesignHeart, 5, false)
	p.AddToDrawPile(c1, c2)
	assert.Equal(t, 2, p.GetDrawPileSize())
}

func TestSpeedPlayer_DrawToHand(t *testing.T) {
	t.Run("draws from pile to hand", func(t *testing.T) {
		p := domain.NewSpeedPlayer(true)
		c := domain.NewCard(domain.CardDesignSpade, 1, false)
		p.AddToDrawPile(c)

		ok := p.DrawToHand()
		assert.True(t, ok)
		assert.Equal(t, 1, p.GetCardsSize())
		assert.Equal(t, 0, p.GetDrawPileSize())
		assert.Equal(t, c, p.GetCard(0))
	})
	t.Run("returns false when empty", func(t *testing.T) {
		p := domain.NewSpeedPlayer(true)
		ok := p.DrawToHand()
		assert.False(t, ok)
		assert.Equal(t, 0, p.GetCardsSize())
	})
}

func TestSpeedPlayer_RefillHand(t *testing.T) {
	p := domain.NewSpeedPlayer(true)
	for i := 1; i <= 10; i++ {
		p.AddToDrawPile(domain.NewCard(domain.CardDesignSpade, i, false))
	}
	p.RefillHand(4)
	assert.Equal(t, 4, p.GetCardsSize())
	assert.Equal(t, 6, p.GetDrawPileSize())
}

func TestSpeedPlayer_RefillHand_NotEnough(t *testing.T) {
	p := domain.NewSpeedPlayer(true)
	p.AddToDrawPile(domain.NewCard(domain.CardDesignSpade, 1, false))
	p.AddToDrawPile(domain.NewCard(domain.CardDesignSpade, 2, false))
	p.RefillHand(4)
	assert.Equal(t, 2, p.GetCardsSize())
	assert.Equal(t, 0, p.GetDrawPileSize())
}

func TestSpeedPlayer_HasCards(t *testing.T) {
	tests := []struct {
		name     string
		hand     int
		draw     int
		expected bool
	}{
		{"both empty", 0, 0, false},
		{"hand only", 1, 0, true},
		{"draw only", 0, 1, true},
		{"both", 1, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := domain.NewSpeedPlayer(true)
			for i := 0; i < tt.hand; i++ {
				p.AddCard(domain.NewCard(domain.CardDesignSpade, i+1, false))
			}
			for i := 0; i < tt.draw; i++ {
				p.AddToDrawPile(domain.NewCard(domain.CardDesignHeart, i+1, false))
			}
			assert.Equal(t, tt.expected, p.HasCards())
		})
	}
}

func TestSpeedPlayer_ResetDrawPile(t *testing.T) {
	p := domain.NewSpeedPlayer(true)
	p.AddToDrawPile(domain.NewCard(domain.CardDesignSpade, 1, false))
	p.ResetDrawPile()
	assert.Equal(t, 0, p.GetDrawPileSize())
}

func TestSpeedPlayer_JSON(t *testing.T) {
	p := domain.NewSpeedPlayer(true)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	p.AddToDrawPile(domain.NewCard(domain.CardDesignHeart, 5, false), domain.NewCard(domain.CardDesignDiamond, 10, false))

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored domain.SpeedPlayer
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.True(t, restored.GetIsHuman())
	assert.Equal(t, 1, restored.GetCardsSize())
	assert.Equal(t, domain.CardDesignSpade, restored.GetCard(0).GetDesign())
	assert.Equal(t, 1, restored.GetCard(0).GetValue())
	assert.Equal(t, 2, restored.GetDrawPileSize())
}

func TestSpeedPlayer_JSON_NilGamePlayer(t *testing.T) {
	data := []byte(`{"dp":[]}`)
	var p domain.SpeedPlayer
	err := json.Unmarshal(data, &p)
	require.NoError(t, err)
	assert.False(t, p.GetIsHuman())
	assert.Equal(t, 0, p.GetCardsSize())
}
