//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertTysiacCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestTysiacDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("discard card index", func(t *testing.T) {
		g := newTestTysiac()
		g.SetDeclarerIdx(0)
		g.SetPhase(domain.TysiacPhaseTalon)
		setTysiacHand(g, 0, tysCard(domain.CardDesignSpade, 1))
		assertTysiacCodedError(t, g.PlayerDiscard(-1), domain.ErrInvalidCard, "tysiac.errCardIndexOutOfRange")
	})

	t.Run("must follow lead suit", func(t *testing.T) {
		g := newTestTysiac()
		g.SetPhase(domain.TysiacPhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.SetTrumpSuit(domain.CardDesignHeart)
		g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: tysCard(domain.CardDesignSpade, 12)}})
		setTysiacHand(g, 0, tysCard(domain.CardDesignSpade, 1), tysCard(domain.CardDesignHeart, 9))
		assertTysiacCodedError(t, g.PlayerPlay(1), domain.ErrInvalidPlay, "tysiac.errFollowLeadSuit")
	})

	t.Run("must play trump when void", func(t *testing.T) {
		g := newTestTysiac()
		g.SetPhase(domain.TysiacPhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.SetTrumpSuit(domain.CardDesignHeart)
		g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: tysCard(domain.CardDesignSpade, 12)}})
		setTysiacHand(g, 0, tysCard(domain.CardDesignHeart, 9), tysCard(domain.CardDesignClover, 10))
		assertTysiacCodedError(t, g.PlayerPlay(1), domain.ErrInvalidPlay, "tysiac.errMustPlayTrump")
	})

	t.Run("must overtrump", func(t *testing.T) {
		g := newTestTysiac()
		g.SetPhase(domain.TysiacPhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.SetTrumpSuit(domain.CardDesignHeart)
		g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: tysCard(domain.CardDesignHeart, 10)}})
		setTysiacHand(g, 0, tysCard(domain.CardDesignHeart, 13), tysCard(domain.CardDesignHeart, 1))
		assertTysiacCodedError(t, g.PlayerPlay(0), domain.ErrInvalidPlay, "tysiac.errMustOvertrump")
	})
}
