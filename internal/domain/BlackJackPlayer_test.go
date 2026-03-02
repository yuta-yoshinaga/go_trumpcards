package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestBlackJackPlayer_Method(t *testing.T) {
	tbp := domain.NewBlackJackPlayer()
	t.Run("success GetScore 0", func(t *testing.T) {
		tbp.Reset()
		assert.Equal(t, 0, tbp.GetScore())
	})
	t.Run("success GetScore 11", func(t *testing.T) {
		tbp.Reset()
		tbp.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		assert.Equal(t, 11, tbp.GetScore())
	})
	t.Run("success GetScore 13", func(t *testing.T) {
		tbp.Reset()
		tbp.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		tbp.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		assert.Equal(t, 13, tbp.GetScore())
	})
	t.Run("success GetScore 13", func(t *testing.T) {
		tbp.Reset()
		tbp.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		tbp.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		tbp.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		assert.Equal(t, 13, tbp.GetScore())
	})
	t.Run("success GetScore 14", func(t *testing.T) {
		tbp.Reset()
		tbp.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		tbp.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		tbp.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		tbp.AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		assert.Equal(t, 14, tbp.GetScore())
	})
	t.Run("success GetScore 23", func(t *testing.T) {
		tbp.Reset()
		tbp.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		tbp.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		tbp.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		tbp.AddCard(domain.NewCard(domain.CardDesignClover, 11, false))
		assert.Equal(t, 23, tbp.GetScore())
	})
}
