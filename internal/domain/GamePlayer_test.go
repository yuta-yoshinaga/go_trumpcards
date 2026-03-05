package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewGamePlayer(t *testing.T) {
	t.Run("human player", func(t *testing.T) {
		gp := domain.NewGamePlayer(true)
		assert.True(t, gp.GetIsHuman())
		assert.False(t, gp.GetIsFinished())
		assert.Equal(t, 0, gp.GetCardsSize())
	})

	t.Run("CPU player", func(t *testing.T) {
		gp := domain.NewGamePlayer(false)
		assert.False(t, gp.GetIsHuman())
		assert.False(t, gp.GetIsFinished())
	})

	t.Run("SetIsFinished", func(t *testing.T) {
		gp := domain.NewGamePlayer(false)
		gp.SetIsFinished(true)
		assert.True(t, gp.GetIsFinished())
		gp.SetIsFinished(false)
		assert.False(t, gp.GetIsFinished())
	})

	t.Run("promoted Player methods work", func(t *testing.T) {
		gp := domain.NewGamePlayer(true)
		gp.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		assert.Equal(t, 1, gp.GetCardsSize())
		assert.Equal(t, 5, gp.GetCard(0).GetValue())
	})
}

func TestNewRankedGamePlayer(t *testing.T) {
	t.Run("human player", func(t *testing.T) {
		rp := domain.NewRankedGamePlayer(true)
		assert.True(t, rp.GetIsHuman())
		assert.False(t, rp.GetIsFinished())
		assert.Equal(t, -1, rp.GetRank())
		assert.Equal(t, 0, rp.GetCardsSize())
	})

	t.Run("CPU player", func(t *testing.T) {
		rp := domain.NewRankedGamePlayer(false)
		assert.False(t, rp.GetIsHuman())
		assert.Equal(t, -1, rp.GetRank())
	})

	t.Run("SetRank", func(t *testing.T) {
		rp := domain.NewRankedGamePlayer(true)
		rp.SetRank(1)
		assert.Equal(t, 1, rp.GetRank())
		rp.SetRank(4)
		assert.Equal(t, 4, rp.GetRank())
	})

	t.Run("SetIsFinished promoted from GamePlayer", func(t *testing.T) {
		rp := domain.NewRankedGamePlayer(false)
		rp.SetIsFinished(true)
		assert.True(t, rp.GetIsFinished())
		rp.SetIsFinished(false)
		assert.False(t, rp.GetIsFinished())
	})

	t.Run("promoted Player methods work", func(t *testing.T) {
		rp := domain.NewRankedGamePlayer(true)
		rp.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		assert.Equal(t, 1, rp.GetCardsSize())
		assert.Equal(t, 10, rp.GetCard(0).GetValue())
	})
}
