package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertKalookiDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func kalookiErrorGame() (*Kalooki, *KalookiPlayer, *KalookiPlayer) {
	g := NewDefaultKalooki()
	g.Reset()
	current := g.GetPlayer(0)
	target := g.GetPlayer(1)
	current.Reset()
	target.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(KalookiPhaseMeld)
	return g, current, target
}

func TestKalooki_AdditionalDomainErrorCallSitesHaveCodes(t *testing.T) {
	t.Run("layoff target meld", func(t *testing.T) {
		g, p, target := kalookiErrorGame()
		p.SetHasOpened(true)
		target.SetHasOpened(true)
		p.AddCard(klCard(CardDesignSpade, 5))
		assertKalookiDomainError(t, g.PlayerLayoff(1, 0, 0), ErrInvalidPlay, "kalooki.errTargetMeldInvalid")
	})

	t.Run("layoff card index", func(t *testing.T) {
		g, p, target := kalookiErrorGame()
		p.SetHasOpened(true)
		target.SetHasOpened(true)
		target.SetMelds([][]*Card{{klCard(CardDesignSpade, 5), klCard(CardDesignHeart, 5), klCard(CardDesignDiamond, 5)}})
		assertKalookiDomainError(t, g.PlayerLayoff(1, 0, -1), ErrInvalidCard, "kalooki.errCardIndexOutOfRange")
	})
}
