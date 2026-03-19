//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewSpadesPlayer(t *testing.T) {
	t.Run("human player", func(t *testing.T) {
		p := domain.NewSpadesPlayer(true)
		assert.True(t, p.GetIsHuman())
		assert.Equal(t, -1, p.GetBid())
		assert.Equal(t, 0, p.GetTrickCount())
		assert.Equal(t, 0, p.GetRoundScore())
		assert.Equal(t, 0, p.GetCumulativeScore())
		assert.Equal(t, 0, p.GetBags())
		assert.Nil(t, p.GetTricksTaken())
	})

	t.Run("CPU player", func(t *testing.T) {
		p := domain.NewSpadesPlayer(false)
		assert.False(t, p.GetIsHuman())
		assert.Equal(t, -1, p.GetBid())
	})
}

func TestSpadesPlayer_BidGetterSetter(t *testing.T) {
	p := domain.NewSpadesPlayer(false)
	p.SetBid(3)
	assert.Equal(t, 3, p.GetBid())
	p.SetBid(0)
	assert.Equal(t, 0, p.GetBid())
}

func TestSpadesPlayer_AddTrick(t *testing.T) {
	p := domain.NewSpadesPlayer(false)
	cards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
	}
	p.AddTrick(cards)
	assert.Equal(t, 1, p.GetTrickCount())
	assert.Len(t, p.GetTricksTaken(), 1)
	assert.Equal(t, cards, p.GetTricksTaken()[0])

	p.AddTrick(cards)
	assert.Equal(t, 2, p.GetTrickCount())
}

func TestSpadesPlayer_RoundScore(t *testing.T) {
	p := domain.NewSpadesPlayer(false)
	p.SetRoundScore(42)
	assert.Equal(t, 42, p.GetRoundScore())
}

func TestSpadesPlayer_CumulativeScore(t *testing.T) {
	p := domain.NewSpadesPlayer(false)
	p.SetCumulativeScore(100)
	assert.Equal(t, 100, p.GetCumulativeScore())
}

func TestSpadesPlayer_Bags(t *testing.T) {
	p := domain.NewSpadesPlayer(false)
	p.SetBags(5)
	assert.Equal(t, 5, p.GetBags())
}

func TestSpadesPlayer_CommitRoundScore(t *testing.T) {
	p := domain.NewSpadesPlayer(false)
	p.SetRoundScore(30)
	p.CommitRoundScore()
	assert.Equal(t, 30, p.GetCumulativeScore())

	p.SetRoundScore(20)
	p.CommitRoundScore()
	assert.Equal(t, 50, p.GetCumulativeScore())
}

func TestSpadesPlayer_ResetRound(t *testing.T) {
	p := domain.NewSpadesPlayer(true)
	p.SetBid(5)
	p.SetRoundScore(30)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	p.AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 2, false)})
	p.SetIsFinished(true)

	p.ResetRound()

	assert.Equal(t, -1, p.GetBid())
	assert.Equal(t, 0, p.GetRoundScore())
	assert.Equal(t, 0, p.GetCardsSize())
	assert.Nil(t, p.GetTricksTaken())
	assert.Equal(t, 0, p.GetTrickCount())
	assert.False(t, p.GetIsFinished())
}
