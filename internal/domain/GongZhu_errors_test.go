//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertGongZhuDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestGongZhuDomainErrorsHaveMessageCodes(t *testing.T) {
	g := newTestGongZhu()
	g.SetPhase(domain.GongZhuPhaseExpose)
	assert.NoError(t, g.PlayerExpose(nil))
	assertGongZhuDomainError(t, g.PlayerExpose(nil), domain.ErrInvalidPlay, "gongzhu.errExposureAlreadySelected")

	g = newTestGongZhu()
	g.SetPhase(domain.GongZhuPhaseExpose)
	g.GetPlayer(0).AddCard(gzCard(domain.CardDesignClover, 2))
	assertGongZhuDomainError(t, g.PlayerExpose([]int{-1}), domain.ErrInvalidCard, "gongzhu.errCardIndexOutOfRange")

	g = newTestGongZhu()
	g.SetPhase(domain.GongZhuPhaseExpose)
	g.GetPlayer(0).AddCard(gzCard(domain.CardDesignSpade, 12))
	assertGongZhuDomainError(t, g.PlayerExpose([]int{0, 0}), domain.ErrInvalidCard, "gongzhu.errDuplicateCardIndex")

	g = newTestGongZhu()
	g.SetPhase(domain.GongZhuPhaseExpose)
	g.GetPlayer(0).AddCard(gzCard(domain.CardDesignClover, 2))
	assertGongZhuDomainError(t, g.PlayerExpose([]int{0}), domain.ErrInvalidPlay, "gongzhu.errOnlyPointCardsExposable")

	g = newTestGongZhu()
	g.SetPhase(domain.GongZhuPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assertGongZhuDomainError(t, g.PlayerPlay(-1), domain.ErrInvalidCard, "gongzhu.errCardIndexOutOfRange")

	g = newTestGongZhu()
	g.SetPhase(domain.GongZhuPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(0).AddCard(gzCard(domain.CardDesignHeart, 2))
	g.GetPlayer(0).AddCard(gzCard(domain.CardDesignClover, 3))
	assertGongZhuDomainError(t, g.PlayerPlay(0), domain.ErrInvalidPlay, "gongzhu.errHeartsNotBroken")

	g = newTestGongZhu()
	g.SetPhase(domain.GongZhuPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: gzCard(domain.CardDesignSpade, 5)}})
	g.GetPlayer(0).AddCard(gzCard(domain.CardDesignHeart, 2))
	g.GetPlayer(0).AddCard(gzCard(domain.CardDesignSpade, 3))
	assertGongZhuDomainError(t, g.PlayerPlay(0), domain.ErrInvalidPlay, "gongzhu.errFollowLeadSuit")
}
