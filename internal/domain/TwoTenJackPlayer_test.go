//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewTwoTenJackPlayer(t *testing.T) {
	t.Run("human", func(t *testing.T) {
		p := domain.NewTwoTenJackPlayer(true)
		assert.True(t, p.GetIsHuman())
		assert.Equal(t, 0, p.GetTrickCount())
		assert.Equal(t, 0, p.GetRoundScore())
		assert.Equal(t, 0, p.GetCumulativeScore())
		assert.Nil(t, p.GetTricksTaken())
	})

	t.Run("cpu", func(t *testing.T) {
		p := domain.NewTwoTenJackPlayer(false)
		assert.False(t, p.GetIsHuman())
	})
}

func TestTwoTenJackPlayer_AddTrick(t *testing.T) {
	p := domain.NewTwoTenJackPlayer(false)
	cards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
	}
	p.AddTrick(cards)
	assert.Equal(t, 1, p.GetTrickCount())
	assert.Len(t, p.GetTricksTaken(), 1)
}

func TestTwoTenJackPlayer_CommitRoundScore(t *testing.T) {
	p := domain.NewTwoTenJackPlayer(false)
	p.SetRoundScore(12)
	p.CommitRoundScore()
	assert.Equal(t, 12, p.GetCumulativeScore())
	p.SetRoundScore(8)
	p.CommitRoundScore()
	assert.Equal(t, 20, p.GetCumulativeScore())
}

func TestTwoTenJackPlayer_ResetRound(t *testing.T) {
	p := domain.NewTwoTenJackPlayer(true)
	p.SetRoundScore(10)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	p.AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 3, false)})
	p.SetIsFinished(true)

	p.ResetRound()

	assert.Equal(t, 0, p.GetRoundScore())
	assert.Equal(t, 0, p.GetCardsSize())
	assert.Nil(t, p.GetTricksTaken())
	assert.Equal(t, 0, p.GetTrickCount())
	assert.False(t, p.GetIsFinished())
}

func TestTwoTenJackPlayer_GetCapturedPointCards(t *testing.T) {
	p := domain.NewTwoTenJackPlayer(false)
	// Trick 1: A(spade)=1, 10(heart)=10, J(club)=1, K(dia)=0 -> 12
	p.AddTrick([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
		domain.NewCard(domain.CardDesignClover, 11, false),
		domain.NewCard(domain.CardDesignDiamond, 13, false),
	})
	// Trick 2: 10(spade)=10, 5(heart)=0 -> 10
	p.AddTrick([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
	})
	assert.Equal(t, 22, p.GetCapturedPointCards())
}

func TestTwoTenJackCardPoints(t *testing.T) {
	assert.Equal(t, 1, domain.TwoTenJackCardPoints(domain.NewCard(domain.CardDesignSpade, 1, false)))
	assert.Equal(t, 10, domain.TwoTenJackCardPoints(domain.NewCard(domain.CardDesignSpade, 10, false)))
	assert.Equal(t, 1, domain.TwoTenJackCardPoints(domain.NewCard(domain.CardDesignSpade, 11, false)))
	assert.Equal(t, 0, domain.TwoTenJackCardPoints(domain.NewCard(domain.CardDesignSpade, 2, false)))
	assert.Equal(t, 0, domain.TwoTenJackCardPoints(domain.NewCard(domain.CardDesignSpade, 12, false)))
	assert.Equal(t, 0, domain.TwoTenJackCardPoints(domain.NewCard(domain.CardDesignSpade, 13, false)))
	assert.Equal(t, 0, domain.TwoTenJackCardPoints(nil))
}
