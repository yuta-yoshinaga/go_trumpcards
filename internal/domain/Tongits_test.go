//go:build test
// +build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func tongitsTestGame() *domain.Tongits {
	players := []*domain.TongitsPlayer{domain.NewTongitsPlayer(true), domain.NewTongitsPlayer(false), domain.NewTongitsPlayer(false)}
	cfg := domain.DefaultTongitsConfig()
	cfg.PointLimit = 10000
	return domain.NewTongits(domain.NewTrumpCards(0), players, cfg)
}
func tongitsCard(suit, value int) *domain.Card { return domain.NewCard(suit, value, false) }

func TestTongitsThreePlayerDeal(t *testing.T) {
	g := tongitsTestGame()
	g.Reset()
	assert.Equal(t, 3, g.GetPlayerCnt())
	assert.Equal(t, 13, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 12, g.GetPlayer(1).GetCardsSize())
	assert.Equal(t, 12, g.GetPlayer(2).GetCardsSize())
	assert.Equal(t, 14, g.GetDrawPileCount())
}

func TestTongitsMeldAndSapawAcrossPlayers(t *testing.T) {
	g := tongitsTestGame()
	g.SetPhase(domain.TongitsPhaseDiscard)
	g.SetCurrentPlayerIdx(0)
	p := g.GetPlayer(0)
	for _, v := range []int{7, 7, 7, 7} {
		p.AddCard(tongitsCard(domain.CardDesignSpade, v))
	}
	require.NoError(t, g.PlayerMeld([]int{0, 1, 2}))
	assert.Len(t, p.GetMelds(), 1)
	require.NoError(t, g.PlayerSapaw(0, 0, 0))
	assert.Len(t, p.GetMeld(0), 4)
}

func TestTongitsSapawRejectsInvalidCard(t *testing.T) {
	g := tongitsTestGame()
	g.SetPhase(domain.TongitsPhaseDiscard)
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(1).AppendMeld([]*domain.Card{tongitsCard(domain.CardDesignHeart, 2), tongitsCard(domain.CardDesignHeart, 3), tongitsCard(domain.CardDesignHeart, 4)})
	g.GetPlayer(0).AddCard(tongitsCard(domain.CardDesignSpade, 9))
	assert.Error(t, g.PlayerSapaw(1, 0, 0))
}

func TestTongitsEmptyHandWinsImmediately(t *testing.T) {
	g := tongitsTestGame()
	g.SetPhase(domain.TongitsPhaseDiscard)
	g.SetCurrentPlayerIdx(0)
	for _, v := range []int{2, 2, 2} {
		g.GetPlayer(0).AddCard(tongitsCard(domain.CardDesignSpade, v))
	}
	require.NoError(t, g.PlayerMeld([]int{0, 1, 2}))
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerIdx())
}

func TestTongitsChallengeComparesThreeHands(t *testing.T) {
	g := tongitsTestGame()
	g.SetPhase(domain.TongitsPhaseDiscard)
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(0).AddCard(tongitsCard(domain.CardDesignSpade, 9))
	g.GetPlayer(1).AddCard(tongitsCard(domain.CardDesignSpade, 2))
	g.GetPlayer(2).AddCard(tongitsCard(domain.CardDesignSpade, 5))
	require.NoError(t, g.PlayerChallenge([]bool{true, true}))
	assert.Equal(t, 1, g.GetWinnerIdx())
}

func TestTongitsConfigAndCardValue(t *testing.T) {
	assert.NoError(t, domain.DefaultTongitsConfig().Validate())
	assert.Equal(t, 1, domain.TongitsCardValue(tongitsCard(domain.CardDesignSpade, 1)))
	assert.Equal(t, 10, domain.TongitsCardValue(tongitsCard(domain.CardDesignSpade, 13)))
}
