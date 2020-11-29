package entities_test

import (
	"testing"
	"go_trumpcards/entities"

	"github.com/stretchr/testify/assert"
)

func TestBlackJack_Method(t *testing.T) {
	tc := entities.NewTrumpCards(0)
	player := entities.NewBlackJackPlayer()
	dealer := entities.NewBlackJackPlayer()
	tb := entities.NewBlackJack(tc, player, dealer)
	t.Run("success Reset", func(t *testing.T) {
		tb.Reset()
	})
	t.Run("success GetGameEndFlag", func(t *testing.T) {
		assert.Equal(t, false, tb.GetGameEndFlag())
	})
	t.Run("success GetPlayer", func(t *testing.T) {
		assert.NotEmpty(t, tb.GetPlayer())
	})
	t.Run("success GetDealer", func(t *testing.T) {
		assert.NotEmpty(t, tb.GetDealer())
	})
	t.Run("success PlayerHit", func(t *testing.T) {
		tb.PlayerHit()
	})
	t.Run("success PlayerStand", func(t *testing.T) {
		tb.PlayerStand()
	})
	t.Run("success DealerHit", func(t *testing.T) {
		tb.DealerHit()
	})
	t.Run("success DealerStand", func(t *testing.T) {
		tb.DealerStand()
	})
	t.Run("success GameJudgment player lose", func(t *testing.T) {
		tb.Reset()
		player.Reset()
		dealer.Reset()
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 2, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 10, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 11, false))
		assert.Equal(t, -1, tb.GameJudgment())
	})
	t.Run("success GameJudgment player lose", func(t *testing.T) {
		tb.Reset()
		player.Reset()
		dealer.Reset()
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 1, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 10, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 11, false))
		assert.Equal(t, -1, tb.GameJudgment())
	})
	t.Run("success GameJudgment player win", func(t *testing.T) {
		tb.Reset()
		player.Reset()
		dealer.Reset()
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 2, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 10, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 11, false))
		assert.Equal(t, 1, tb.GameJudgment())
	})
	t.Run("success GameJudgment draw", func(t *testing.T) {
		tb.Reset()
		player.Reset()
		dealer.Reset()
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 1, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 10, false))
		assert.Equal(t, 0, tb.GameJudgment())
	})
	t.Run("success GameJudgment player win", func(t *testing.T) {
		tb.Reset()
		player.Reset()
		dealer.Reset()
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 1, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 10, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 11, false))
		assert.Equal(t, 1, tb.GameJudgment())
	})
	t.Run("success GameJudgment player win", func(t *testing.T) {
		tb.Reset()
		player.Reset()
		dealer.Reset()
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 10, false))
		assert.Equal(t, 1, tb.GameJudgment())
	})
	t.Run("success GameJudgment player lose", func(t *testing.T) {
		tb.Reset()
		player.Reset()
		dealer.Reset()
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 1, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 10, false))
		assert.Equal(t, -1, tb.GameJudgment())
	})
}
