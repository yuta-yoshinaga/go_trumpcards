//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertCallBreakDomainError(t *testing.T, err error, sentinel error, code string, params map[string]string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Equal(t, params, de.MessageParams())
}

func TestCallBreakDomainErrorsHaveMessageCodes(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*domain.CallBreak)
		play     int
		code     string
		params   map[string]string
		sentinel error
	}{
		{name: "bid range", setup: func(cb *domain.CallBreak) { setupCallBreakErrorBid(cb, 0) }, play: 0, code: "callbreak.errBidRange", params: map[string]string{"min": "1", "max": "13"}, sentinel: domain.ErrInvalidPlay},
		{name: "card index out of range", setup: func(cb *domain.CallBreak) { setupCallBreakErrorPlay(cb, 0, nil) }, play: -1, code: "callbreak.errCardIndexOutOfRange", sentinel: domain.ErrInvalidCard},
		{
			name: "spades not broken",
			setup: func(cb *domain.CallBreak) {
				setupCallBreakErrorPlay(cb, 0, []*domain.TrickCard{})
				setCallBreakHand(cb, 0, callBreakCard(domain.CardDesignSpade, 2), callBreakCard(domain.CardDesignHeart, 3))
			}, play: 0, code: "callbreak.errSpadesNotBroken", sentinel: domain.ErrInvalidPlay,
		},
		{
			name: "follow lead suit",
			setup: func(cb *domain.CallBreak) {
				setupCallBreakErrorPlay(cb, 0, []*domain.TrickCard{{PlayerIdx: 1, Card: callBreakCard(domain.CardDesignHeart, 9)}})
				setCallBreakHand(cb, 0, callBreakCard(domain.CardDesignHeart, 2), callBreakCard(domain.CardDesignSpade, 3))
			}, play: 1, code: "callbreak.errFollowLeadSuit", sentinel: domain.ErrInvalidPlay,
		},
		{
			name: "must trump",
			setup: func(cb *domain.CallBreak) {
				setupCallBreakErrorPlay(cb, 0, []*domain.TrickCard{{PlayerIdx: 1, Card: callBreakCard(domain.CardDesignHeart, 9)}})
				setCallBreakHand(cb, 0, callBreakCard(domain.CardDesignSpade, 2), callBreakCard(domain.CardDesignClover, 3))
			}, play: 1, code: "callbreak.errMustTrump", sentinel: domain.ErrInvalidPlay,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := newTestCallBreak()
			tt.setup(cb)
			var err error
			if tt.name == "bid range" {
				err = cb.PlayerBid(0)
			} else {
				err = cb.PlayerPlay(tt.play)
			}
			assertCallBreakDomainError(t, err, tt.sentinel, tt.code, tt.params)
		})
	}
}

func setupCallBreakErrorBid(cb *domain.CallBreak, idx int) {
	cb.SetPhase(domain.CallBreakPhaseBid)
	cb.SetBidPlayerIdx(idx)
}
func setupCallBreakErrorPlay(cb *domain.CallBreak, current int, trick []*domain.TrickCard) {
	cb.SetPhase(domain.CallBreakPhasePlay)
	cb.SetCurrentPlayerIdx(current)
	cb.SetCurrentTrick(trick)
	cb.SetSpadesBroken(false)
}
func setCallBreakHand(cb *domain.CallBreak, idx int, cards ...*domain.Card) {
	p := cb.GetPlayer(idx)
	p.Reset()
	for _, card := range cards {
		p.AddCard(card)
	}
}
func callBreakCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }
