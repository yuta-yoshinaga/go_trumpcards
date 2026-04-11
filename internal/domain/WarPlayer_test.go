//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWarPlayer(t *testing.T) {
	p := NewWarPlayer(true)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, 0, p.GetDrawPileSize())
	assert.Equal(t, 0, p.GetDiscardPileSize())
	assert.Equal(t, 0, p.TotalCards())
	assert.False(t, p.HasCards())
}

func TestWarPlayer_AddAndDraw(t *testing.T) {
	p := NewWarPlayer(false)
	c1 := NewCard(CardDesignSpade, 5, false)
	c2 := NewCard(CardDesignHeart, 7, false)
	p.AddToDrawPile(c1, c2)

	assert.Equal(t, 2, p.GetDrawPileSize())
	assert.True(t, p.HasCards())

	drawn := p.DrawOne()
	assert.Same(t, c1, drawn)
	assert.Equal(t, 1, p.GetDrawPileSize())

	drawn = p.DrawOne()
	assert.Same(t, c2, drawn)
	assert.Equal(t, 0, p.GetDrawPileSize())

	assert.Nil(t, p.DrawOne())
}

func TestWarPlayer_RefillFromDiscard(t *testing.T) {
	p := NewWarPlayer(false)
	c1 := NewCard(CardDesignSpade, 5, false)
	c2 := NewCard(CardDesignHeart, 7, false)
	c3 := NewCard(CardDesignClover, 9, false)
	p.AddToDiscardPile(c1, c2, c3)

	// draw pile empty but discard has 3 -> refill and draw
	assert.Equal(t, 3, p.TotalCards())
	drawn := p.DrawOne()
	assert.NotNil(t, drawn)
	assert.Equal(t, 2, p.TotalCards())
	assert.Equal(t, 0, p.GetDiscardPileSize())
	assert.Equal(t, 2, p.GetDrawPileSize())
}

func TestWarPlayer_ResetPiles(t *testing.T) {
	p := NewWarPlayer(false)
	p.AddToDrawPile(NewCard(CardDesignSpade, 1, false))
	p.AddToDiscardPile(NewCard(CardDesignHeart, 2, false))
	p.ResetPiles()
	assert.Equal(t, 0, p.TotalCards())
}

func TestWarPlayer_JSON(t *testing.T) {
	p := NewWarPlayer(true)
	p.AddToDrawPile(NewCard(CardDesignSpade, 5, false))
	p.AddToDiscardPile(NewCard(CardDesignHeart, 7, false))

	data, err := json.Marshal(p)
	assert.NoError(t, err)

	var decoded WarPlayer
	assert.NoError(t, json.Unmarshal(data, &decoded))
	assert.True(t, decoded.GetIsHuman())
	assert.Equal(t, 1, decoded.GetDrawPileSize())
	assert.Equal(t, 1, decoded.GetDiscardPileSize())
}
