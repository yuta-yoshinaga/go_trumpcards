//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertKlaverjasDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel))
	de, ok := err.(*DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestKlaverjasDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("card index out of range", func(t *testing.T) {
		g := newKlavGame(true)
		g.SetPhase(KlaverjasPhasePlay)
		g.SetCurrentPlayerIdx(0)
		assertKlaverjasDomainError(t, g.PlayerPlay(-1), ErrInvalidCard, "klaverjas.errCardIndexOutOfRange")
	})

	t.Run("must follow lead suit", func(t *testing.T) {
		g := newKlavGame(true)
		g.SetPhase(KlaverjasPhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.SetTrumpSuit(CardDesignDiamond)
		g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 3, Card: klavCard(CardDesignClover, 1)}})
		klavSetHand(g.GetPlayer(0), klavCard(CardDesignClover, 13), klavCard(CardDesignDiamond, 7))
		assertKlaverjasDomainError(t, g.PlayerPlay(1), ErrInvalidPlay, "klaverjas.errFollowLeadSuit")
	})

	t.Run("must play trump when void", func(t *testing.T) {
		g := newKlavGame(true)
		g.SetPhase(KlaverjasPhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.SetTrumpSuit(CardDesignDiamond)
		g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 3, Card: klavCard(CardDesignClover, 1)}})
		klavSetHand(g.GetPlayer(0), klavCard(CardDesignDiamond, 7), klavCard(CardDesignHeart, 1))
		assertKlaverjasDomainError(t, g.PlayerPlay(1), ErrInvalidPlay, "klaverjas.errMustPlayTrump")
	})

	t.Run("must overtrump", func(t *testing.T) {
		g := newKlavGame(true)
		g.SetPhase(KlaverjasPhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.SetTrumpSuit(CardDesignDiamond)
		g.SetCurrentTrick([]*TrickCard{
			{PlayerIdx: 3, Card: klavCard(CardDesignClover, 1)},
			{PlayerIdx: 1, Card: klavCard(CardDesignDiamond, 9)},
		})
		klavSetHand(g.GetPlayer(0), klavCard(CardDesignDiamond, 11), klavCard(CardDesignDiamond, 7))
		assertKlaverjasDomainError(t, g.PlayerPlay(1), ErrInvalidPlay, "klaverjas.errMustOvertrump")
	})
}
