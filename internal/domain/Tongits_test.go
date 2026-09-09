//go:build test
// +build test

package domain_test

import (
	"encoding/json"
	"strings"
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

func TestTongitsHardCPUChallengeFinishesRound(t *testing.T) {
	g := tongitsTestGame()
	cfg := g.GetConfig()
	cfg.CpuDifficulty = domain.TongitsCpuDifficultyHard
	g.SetConfig(cfg)
	g.SetPhase(domain.TongitsPhaseDiscard)
	g.SetCurrentPlayerIdx(1)
	g.GetPlayer(1).AddCard(tongitsCard(domain.CardDesignSpade, 4))
	g.GetPlayer(0).AddCard(tongitsCard(domain.CardDesignHeart, 9))
	g.GetPlayer(2).AddCard(tongitsCard(domain.CardDesignClover, 8))

	g.CpuPlay()

	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetWinnerIdx())
}

func TestTongitsHardCPUDiscardsWhenChallengeDoesNotApply(t *testing.T) {
	g := tongitsTestGame()
	cfg := g.GetConfig()
	cfg.CpuDifficulty = domain.TongitsCpuDifficultyHard
	g.SetConfig(cfg)
	g.SetPhase(domain.TongitsPhaseDiscard)
	g.SetCurrentPlayerIdx(1)
	g.GetPlayer(1).AddCard(tongitsCard(domain.CardDesignSpade, 4))
	g.GetPlayer(1).AddCard(tongitsCard(domain.CardDesignHeart, 8))

	g.CpuPlay()

	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
	assert.Equal(t, domain.TongitsPhaseDraw, g.GetPhase())
	assert.Len(t, g.GetDiscardPile(), 1)
}

func TestTongitsPlayerDrawAndDiscard(t *testing.T) {
	t.Run("stock", func(t *testing.T) {
		g := tongitsTestGame()
		g.SetPhase(domain.TongitsPhaseDraw)
		g.SetCurrentPlayerIdx(0)
		g.SetDrawPile([]*domain.Card{tongitsCard(domain.CardDesignSpade, 4)})

		require.NoError(t, g.PlayerDrawFromStock())
		assert.Equal(t, domain.TongitsPhaseDiscard, g.GetPhase())
		assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
		assert.Equal(t, 0, g.GetDrawPileCount())

		require.NoError(t, g.PlayerDiscard(0))
		assert.Equal(t, domain.TongitsPhaseDraw, g.GetPhase())
		assert.Len(t, g.GetDiscardPile(), 1)
	})

	t.Run("discard pile", func(t *testing.T) {
		g := tongitsTestGame()
		g.SetPhase(domain.TongitsPhaseDraw)
		g.SetCurrentPlayerIdx(0)
		g.SetDiscardPile([]*domain.Card{tongitsCard(domain.CardDesignHeart, 5)})

		require.NoError(t, g.PlayerDrawFromDiscard())
		assert.Equal(t, domain.TongitsPhaseDiscard, g.GetPhase())
		assert.Empty(t, g.GetDiscardPile())
	})
}

func TestTongitsPlayerDrawFromEmptyStockEndsRound(t *testing.T) {
	g := tongitsTestGame()
	g.SetPhase(domain.TongitsPhaseDraw)
	g.SetCurrentPlayerIdx(0)
	g.SetDrawPile(nil)

	require.NoError(t, g.PlayerDrawFromStock())
	assert.Equal(t, domain.TongitsPhaseRoundEnd, g.GetPhase())
}

func TestTongitsPlayerMethodsRejectInvalidState(t *testing.T) {
	methods := []struct {
		name  string
		phase domain.TongitsPhase
		call  func(*domain.Tongits) error
	}{
		{name: "draw stock", phase: domain.TongitsPhaseDraw, call: func(g *domain.Tongits) error { return g.PlayerDrawFromStock() }},
		{name: "draw discard", phase: domain.TongitsPhaseDraw, call: func(g *domain.Tongits) error { return g.PlayerDrawFromDiscard() }},
		{name: "discard", phase: domain.TongitsPhaseDiscard, call: func(g *domain.Tongits) error { return g.PlayerDiscard(0) }},
		{name: "challenge", phase: domain.TongitsPhaseDiscard, call: func(g *domain.Tongits) error { return g.PlayerChallenge([]bool{true, true}) }},
	}

	for _, tt := range methods {
		t.Run(tt.name+" wrong phase", func(t *testing.T) {
			g := tongitsTestGame()
			wrongPhase := domain.TongitsPhaseDiscard
			if tt.phase == domain.TongitsPhaseDiscard {
				wrongPhase = domain.TongitsPhaseDraw
			}
			g.SetPhase(wrongPhase)
			g.SetCurrentPlayerIdx(0)
			assert.ErrorIs(t, tt.call(g), domain.ErrWrongPhase)
		})
		t.Run(tt.name+" CPU turn", func(t *testing.T) {
			g := tongitsTestGame()
			g.SetPhase(tt.phase)
			g.SetCurrentPlayerIdx(1)
			assert.ErrorIs(t, tt.call(g), domain.ErrNotHumanTurn)
		})
	}

	g := tongitsTestGame()
	g.SetPhase(domain.TongitsPhaseDiscard)
	g.SetCurrentPlayerIdx(0)
	for _, v := range []int{2, 2, 2} {
		g.GetPlayer(0).AddCard(tongitsCard(domain.CardDesignSpade, v))
	}
	require.NoError(t, g.PlayerMeld([]int{0, 1, 2}))
	assert.ErrorIs(t, g.PlayerDiscard(0), domain.ErrGameEnded)
}

func TestTongitsPlayerDiscardRejectsInvalidIndexAndChallengeInput(t *testing.T) {
	g := tongitsTestGame()
	g.SetPhase(domain.TongitsPhaseDiscard)
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(0).AddCard(tongitsCard(domain.CardDesignSpade, 5))

	assert.ErrorIs(t, g.PlayerDiscard(-1), domain.ErrInvalidCard)
	assert.ErrorIs(t, g.PlayerDiscard(1), domain.ErrInvalidCard)
	assert.ErrorIs(t, g.PlayerChallenge([]bool{true}), domain.ErrInvalidPlay)

	require.NoError(t, g.PlayerChallenge([]bool{false, true}))
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	assert.Equal(t, domain.TongitsPhaseDraw, g.GetPhase())
}

func TestTongitsPlayerDrawFromDiscardRejectsEmptyPile(t *testing.T) {
	g := tongitsTestGame()
	g.SetPhase(domain.TongitsPhaseDraw)
	g.SetCurrentPlayerIdx(0)

	assert.ErrorIs(t, g.PlayerDrawFromDiscard(), domain.ErrInvalidPlay)
}

func TestTongitsNextRoundResetsRoundState(t *testing.T) {
	g := tongitsTestGame()
	g.SetRoundNumber(3)
	g.SetPhase(domain.TongitsPhaseRoundEnd)
	g.SetCurrentPlayerIdx(2)
	g.SetDiscardPile([]*domain.Card{tongitsCard(domain.CardDesignSpade, 2)})

	g.NextRound()

	assert.Equal(t, 4, g.GetRoundNumber())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.Len(t, g.GetDiscardPile(), 1)
}

func TestTongitsNextRoundIgnoresWrongPhase(t *testing.T) {
	g := tongitsTestGame()
	g.SetRoundNumber(3)
	g.SetPhase(domain.TongitsPhaseDraw)
	g.NextRound()
	assert.Equal(t, 3, g.GetRoundNumber())
}

func TestTongitsJSONRoundTripAndInvalidInput(t *testing.T) {
	g := tongitsTestGame()
	g.SetPhase(domain.TongitsPhaseDiscard)
	g.SetCurrentPlayerIdx(1)
	g.SetDrawPile([]*domain.Card{tongitsCard(domain.CardDesignSpade, 2)})
	g.SetDiscardPile([]*domain.Card{tongitsCard(domain.CardDesignHeart, 3)})
	g.SetIsTongits(true)

	data, err := json.Marshal(g)
	require.NoError(t, err)
	var restored domain.Tongits
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, domain.TongitsPhaseDiscard, restored.GetPhase())
	assert.Equal(t, 1, restored.GetCurrentPlayerIdx())
	assert.Equal(t, 1, restored.GetDrawPileCount())
	assert.Len(t, restored.GetDiscardPile(), 1)
	assert.True(t, restored.GetIsTongits())

	assert.Error(t, json.Unmarshal([]byte("{"), &restored))
	oversizedPlayers := `{"pl":[` + strings.Repeat("null,", 1000) + `null]}`
	assert.Error(t, json.Unmarshal([]byte(oversizedPlayers), &restored))
}
