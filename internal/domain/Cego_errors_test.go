//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertCegoCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func cegoErrorExchange() *domain.Cego {
	g := domain.NewDefaultCego()
	g.SetDeclarerIdx(0)
	g.SetPhase(domain.CegoPhaseExchange)
	cegoSetHand(g, 0,
		cegoSuitCard(domain.CardDesignSpade, 1), cegoSuitCard(domain.CardDesignSpade, 2),
		cegoSuitCard(domain.CardDesignHeart, 1), cegoSuitCard(domain.CardDesignHeart, 2),
		cegoSuitCard(domain.CardDesignDiamond, 1), cegoSuitCard(domain.CardDesignDiamond, 2),
		cegoSuitCard(domain.CardDesignClover, 1), cegoSuitCard(domain.CardDesignClover, 2),
		cegoSuitCard(domain.CardDesignClover, 3), cegoSuitCard(domain.CardDesignClover, 4),
		cegoTrumpCard(3))
	return g
}

func TestCegoDomainErrorsHaveMessageCodes(t *testing.T) {
	g := domain.NewDefaultCego()
	g.SetBidPlayerIdx(0)
	assertCegoCodedError(t, g.PlayerBid(domain.CegoBid(99)), domain.ErrInvalidPlay, "cego.errInvalidBid")

	g = domain.NewDefaultCego()
	g.SetBidPlayerIdx(0)
	g.SetHighestBid(domain.CegoBidPlay)
	assertCegoCodedError(t, g.PlayerBid(domain.CegoBidPlay), domain.ErrInvalidPlay, "cego.errHigherBidRequired")

	g = domain.NewDefaultCego()
	g.SetDeclarerIdx(0)
	g.SetPhase(domain.CegoPhaseContract)
	assertCegoCodedError(t, g.PlayerChooseContract(domain.CegoContract(99)), domain.ErrInvalidPlay, "cego.errInvalidContract")

	g = cegoErrorExchange()
	assertCegoCodedError(t, g.PlayerDiscard(nil), domain.ErrInvalidCard, "cego.errKeepOneCard")
	g = cegoErrorExchange()
	assertCegoCodedError(t, g.PlayerDiscard([]int{99}), domain.ErrInvalidCard, "cego.errCardIndexOutOfRange")
	g = cegoErrorExchange()
	assertCegoCodedError(t, g.PlayerDiscard([]int{0, 0}), domain.ErrInvalidCard, "cego.errKeepOneCard")
	// 件数検査は範囲検査より先に走る (枚数も範囲も誤っている入力で件数エラーが返る)。
	assertCegoCodedError(t, g.PlayerDiscard([]int{99, 99}), domain.ErrInvalidCard, "cego.errKeepOneCard")

	g = domain.NewDefaultCego()
	g.SetPhase(domain.CegoPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assertCegoCodedError(t, g.PlayerPlay(-1), domain.ErrInvalidCard, "cego.errCardIndexOutOfRange")
}
