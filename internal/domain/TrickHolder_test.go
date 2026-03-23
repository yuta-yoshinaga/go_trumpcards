//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrickHolder_GetTricksTaken(t *testing.T) {
	h := &TrickHolder{}
	assert.Nil(t, h.GetTricksTaken())
	assert.Equal(t, 0, h.GetTrickCount())
}

func TestTrickHolder_AddTrick(t *testing.T) {
	h := &TrickHolder{}
	cards1 := []*Card{NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 2, false)}
	cards2 := []*Card{NewCard(CardDesignDiamond, 3, false)}

	h.AddTrick(cards1)
	assert.Equal(t, 1, h.GetTrickCount())
	assert.Equal(t, cards1, h.GetTricksTaken()[0])

	h.AddTrick(cards2)
	assert.Equal(t, 2, h.GetTrickCount())
	assert.Equal(t, cards2, h.GetTricksTaken()[1])
}
