package entities_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"

	"github.com/stretchr/testify/assert"
)

func TestBlackJackPlayer_Method(t *testing.T) {
	tbp := entities.NewBlackJackPlayer()
	t.Run("success GetScore 0", func(t *testing.T) {
		tbp.Reset()
		assert.Equal(t, 0, tbp.GetScore())
	})
	t.Run("success GetScore 11", func(t *testing.T) {
		tbp.Reset()
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		assert.Equal(t, 11, tbp.GetScore())
	})
	t.Run("success GetScore 13", func(t *testing.T) {
		tbp.Reset()
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		assert.Equal(t, 13, tbp.GetScore())
	})
	t.Run("success GetScore 13", func(t *testing.T) {
		tbp.Reset()
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		assert.Equal(t, 13, tbp.GetScore())
	})
	t.Run("success GetScore 14", func(t *testing.T) {
		tbp.Reset()
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		tbp.AddCard(entities.NewCard(entities.CardDesignClover, 1, false))
		assert.Equal(t, 14, tbp.GetScore())
	})
	t.Run("success GetScore 23", func(t *testing.T) {
		tbp.Reset()
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		tbp.AddCard(entities.NewCard(entities.CardDesignClover, 11, false))
		assert.Equal(t, 23, tbp.GetScore())
	})
}

func TestBlackJackPlayer_Chips(t *testing.T) {
	tbp := entities.NewBlackJackPlayer()
	t.Run("initial chips is 0", func(t *testing.T) {
		assert.Equal(t, 0, tbp.GetChips())
	})
	t.Run("SetChips", func(t *testing.T) {
		tbp.SetChips(1000)
		assert.Equal(t, 1000, tbp.GetChips())
	})
	t.Run("AddChips", func(t *testing.T) {
		tbp.AddChips(500)
		assert.Equal(t, 1500, tbp.GetChips())
	})
	t.Run("SubtractChips success", func(t *testing.T) {
		ok := tbp.SubtractChips(300)
		assert.True(t, ok)
		assert.Equal(t, 1200, tbp.GetChips())
	})
	t.Run("SubtractChips insufficient", func(t *testing.T) {
		ok := tbp.SubtractChips(2000)
		assert.False(t, ok)
		assert.Equal(t, 1200, tbp.GetChips())
	})
}
