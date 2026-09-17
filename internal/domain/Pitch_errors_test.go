//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertPitchDomainError(t *testing.T, err error, sentinel error, code string, params map[string]string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Equal(t, params, de.MessageParams())
}

func TestPitchDomainErrorsHaveMessageCodes(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetDealerIdx(0)
	p.SetBidPlayerIdx(0)
	p.GetPlayer(1).SetBid(domain.PitchPassBid)
	p.GetPlayer(2).SetBid(domain.PitchPassBid)
	p.GetPlayer(3).SetBid(domain.PitchPassBid)
	assertPitchDomainError(t, p.PlayerBid(domain.PitchPassBid), domain.ErrInvalidPlay, "pitch.errDealerCannotPass", nil)

	p.Reset()
	assertPitchDomainError(t, p.PlayerBid(1), domain.ErrInvalidPlay, "pitch.errBidRange", map[string]string{"min": "2", "max": "4"})

	p.Reset()
	p.SetCurrentBid(3)
	assertPitchDomainError(t, p.PlayerBid(3), domain.ErrInvalidPlay, "pitch.errBidMustExceedCurrent", map[string]string{"bid": "3"})

	p.Reset()
	p.SetPhase(domain.PitchPhasePlay)
	p.SetCurrentPlayerIdx(0)
	assertPitchDomainError(t, p.PlayerPlay(-1), domain.ErrInvalidCard, "pitch.errCardIndexOutOfRange", nil)

	p.Reset()
	p.SetPhase(domain.PitchPhasePlay)
	p.SetCurrentPlayerIdx(0)
	p.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: newPitchCard(domain.CardDesignClover, 2)}})
	setHandPitch(p.GetPlayer(0), newPitchCard(domain.CardDesignClover, 5), newPitchCard(domain.CardDesignHeart, 5))
	assertPitchDomainError(t, p.PlayerPlay(1), domain.ErrInvalidPlay, "pitch.errFollowLeadSuitOrTrump", nil)
}
