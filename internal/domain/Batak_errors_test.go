//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertBatakDomainError(t *testing.T, err error, sentinel error, code string, params map[string]string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Equal(t, params, de.MessageParams())
}

func TestBatakDomainErrorsHaveMessageCodes(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	setupBatakBid(cb, 0)
	assertBatakDomainError(t, cb.PlayerBid(4), domain.ErrInvalidPlay, "batak.errBidRange", map[string]string{
		"pass": "0", "min": "5", "max": "13",
	})

	cb.Reset()
	setupBatakPlay(cb, 0, 0, 1)
	assertBatakDomainError(t, cb.PlayerPlay(-1), domain.ErrInvalidCard, "batak.errCardIndexOutOfRange", nil)

	cb.Reset()
	setupBatakPlay(cb, 0, 0, 1)
	cb.GetPlayer(0).Reset()
	cb.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	cb.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	cb.SetSpadesBroken(false)
	assertBatakDomainError(t, cb.PlayerPlay(0), domain.ErrInvalidPlay, "batak.errSpadesNotBroken", nil)

	cb.Reset()
	setupBatakPlay(cb, 0, 0, 1)
	cb.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 2, false)}})
	cb.GetPlayer(0).Reset()
	cb.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	cb.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	assertBatakDomainError(t, cb.PlayerPlay(1), domain.ErrInvalidPlay, "batak.errFollowLeadSuit", nil)

	cb.Reset()
	setupBatakPlay(cb, 0, 0, 1)
	cb.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 2, false)}})
	cb.GetPlayer(0).Reset()
	cb.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	cb.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	assertBatakDomainError(t, cb.PlayerPlay(1), domain.ErrInvalidPlay, "batak.errMustTrumpWhenVoid", nil)
}
