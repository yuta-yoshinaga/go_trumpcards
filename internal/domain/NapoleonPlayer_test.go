//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewNapoleonPlayer(t *testing.T) {
	t.Run("human player", func(t *testing.T) {
		p := domain.NewNapoleonPlayer(true)
		assert.True(t, p.GetIsHuman())
		assert.Equal(t, -1, p.GetBid())
		assert.False(t, p.GetIsNapoleon())
		assert.False(t, p.GetIsAdjutant())
		assert.False(t, p.GetAdjutantRevealed())
		assert.Equal(t, 0, p.GetPictureCards())
		assert.Equal(t, 0, p.GetTrickCount())
		assert.Equal(t, 0, p.GetRoundScore())
		assert.Equal(t, 0, p.GetCumulativeScore())
	})

	t.Run("CPU player", func(t *testing.T) {
		p := domain.NewNapoleonPlayer(false)
		assert.False(t, p.GetIsHuman())
		assert.Equal(t, -1, p.GetBid())
	})
}

func TestNapoleonPlayer_Setters(t *testing.T) {
	p := domain.NewNapoleonPlayer(true)

	p.SetBid(13)
	assert.Equal(t, 13, p.GetBid())

	p.SetIsNapoleon(true)
	assert.True(t, p.GetIsNapoleon())

	p.SetIsAdjutant(true)
	assert.True(t, p.GetIsAdjutant())

	p.SetAdjutantRevealed(true)
	assert.True(t, p.GetAdjutantRevealed())

	p.SetPictureCards(5)
	assert.Equal(t, 5, p.GetPictureCards())

	p.SetRoundScore(10)
	assert.Equal(t, 10, p.GetRoundScore())

	p.SetCumulativeScore(50)
	assert.Equal(t, 50, p.GetCumulativeScore())
}

func TestNapoleonPlayer_AddTrick(t *testing.T) {
	p := domain.NewNapoleonPlayer(true)
	cards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignHeart, 13, false),
	}
	p.AddTrick(cards)
	assert.Equal(t, 1, p.GetTrickCount())
	assert.Len(t, p.GetTricksTaken(), 1)
	assert.Equal(t, cards, p.GetTricksTaken()[0])

	p.AddTrick(cards)
	assert.Equal(t, 2, p.GetTrickCount())
}

func TestNapoleonPlayer_CommitRoundScore(t *testing.T) {
	p := domain.NewNapoleonPlayer(true)
	p.SetRoundScore(12)
	p.CommitRoundScore()
	assert.Equal(t, 12, p.GetCumulativeScore())
	assert.Equal(t, 12, p.GetRoundScore())

	p.SetRoundScore(-5)
	p.CommitRoundScore()
	assert.Equal(t, 7, p.GetCumulativeScore())
}

func TestNapoleonPlayer_ResetRound(t *testing.T) {
	p := domain.NewNapoleonPlayer(true)
	p.SetBid(13)
	p.SetIsNapoleon(true)
	p.SetIsAdjutant(true)
	p.SetAdjutantRevealed(true)
	p.SetPictureCards(5)
	p.SetRoundScore(10)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	p.AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 1, false)})

	p.ResetRound()

	assert.Equal(t, -1, p.GetBid())
	assert.False(t, p.GetIsNapoleon())
	assert.False(t, p.GetIsAdjutant())
	assert.False(t, p.GetAdjutantRevealed())
	assert.Equal(t, 0, p.GetPictureCards())
	assert.Equal(t, 0, p.GetRoundScore())
	assert.Equal(t, 0, p.GetTrickCount())
	assert.Equal(t, 0, p.GetCardsSize())
	assert.False(t, p.GetIsFinished())
}
