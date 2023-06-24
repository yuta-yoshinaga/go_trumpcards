package entities_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"

	"github.com/stretchr/testify/assert"
)

func TestBlackJackPlayer_Method(t *testing.T) {
	tbp := entities.NewBlackJackPlayer()
	t.Run("success CalcScore 0", func(t *testing.T) {
		tbp.Reset()
		tbp.CalcScore()
		assert.Equal(t, 0, tbp.GetScore())
	})
	t.Run("success CalcScore 11", func(t *testing.T) {
		tbp.Reset()
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		tbp.CalcScore()
		assert.Equal(t, 11, tbp.GetScore())
	})
	t.Run("success CalcScore 13", func(t *testing.T) {
		tbp.Reset()
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		tbp.CalcScore()
		assert.Equal(t, 13, tbp.GetScore())
	})
	t.Run("success CalcScore 13", func(t *testing.T) {
		tbp.Reset()
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		tbp.CalcScore()
		assert.Equal(t, 13, tbp.GetScore())
	})
	t.Run("success CalcScore 14", func(t *testing.T) {
		tbp.Reset()
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		tbp.AddCard(entities.NewCard(entities.CardDesignClover, 1, false))
		tbp.CalcScore()
		assert.Equal(t, 14, tbp.GetScore())
	})
	t.Run("success CalcScore 23", func(t *testing.T) {
		tbp.Reset()
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		tbp.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		tbp.AddCard(entities.NewCard(entities.CardDesignClover, 11, false))
		tbp.CalcScore()
		assert.Equal(t, 23, tbp.GetScore())
	})
}
