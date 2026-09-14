//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertCoincheDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func TestCoincheBidAndDoubleErrorsHaveMessageCodes(t *testing.T) {
	g := domain.NewDefaultCoinche()
	g.Reset()
	g.SetBidPlayerIdx(0)
	assertCoincheDomainError(t, g.PlayerBid(100, 0), domain.ErrInvalidPlay, "coinche.errTrumpSuitRequired")
	assertCoincheDomainError(t, g.PlayerBid(1, domain.CardDesignSpade), domain.ErrInvalidPlay, "coinche.errBidRange")
	g.SetContractPoints(100)
	assertCoincheDomainError(t, g.PlayerBid(100, domain.CardDesignSpade), domain.ErrInvalidPlay, "coinche.errBidHigherThanHighest")

	g.SetPhase(domain.CoinchePhaseDouble)
	g.SetCurrentPlayerIdx(0)
	g.SetMakerTeam(0)
	g.SetDouble(domain.CoincheDoubleNone)
	assertCoincheDomainError(t, g.PlayerCoinche(), domain.ErrInvalidPlay, "coinche.errCoincheNotAllowed")
	g.SetMakerTeam(1)
	g.SetDouble(domain.CoincheDoubleCoinche)
	assertCoincheDomainError(t, g.PlayerSurcoinche(), domain.ErrInvalidPlay, "coinche.errSurcoincheNotAllowed")
}

func TestCoinchePlayErrorsHaveMessageCodes(t *testing.T) {
	g := domain.NewDefaultCoinche()
	g.Reset()
	g.SetPhase(domain.CoinchePhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(0).ResetRound()
	assertCoincheDomainError(t, g.PlayerPlay(-1), domain.ErrInvalidCard, "coinche.errCardIndexOutOfRange")

	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 7, false)}})
	g.SetTrumpSuit(domain.CardDesignDiamond)
	assertCoincheDomainError(t, g.PlayerPlay(0), domain.ErrInvalidPlay, "coinche.errFollowLeadSuit")

	g.GetPlayer(0).ResetRound()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
	assertCoincheDomainError(t, g.PlayerPlay(0), domain.ErrInvalidPlay, "coinche.errMustPlayTrump")

	g.GetPlayer(0).ResetRound()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 10, false)}})
	g.SetTrumpSuit(domain.CardDesignSpade)
	assertCoincheDomainError(t, g.PlayerPlay(0), domain.ErrInvalidPlay, "coinche.errMustOvertrump")

	// リードが切り札で切り札を持っているのに、切り札でない札を出した場合。
	// 上の errMustOvertrump とは「出した札が切り札かどうか」だけが違う。
	g.GetPlayer(0).ResetRound()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 10, false)}})
	g.SetTrumpSuit(domain.CardDesignSpade)
	assertCoincheDomainError(t, g.PlayerPlay(1), domain.ErrInvalidPlay, "coinche.errMustFollowTrump")
}
