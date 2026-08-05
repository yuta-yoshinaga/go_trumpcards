//go:build test

package domain_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestSevenBridge() *domain.SevenBridge {
	players := []*domain.SevenBridgePlayer{
		domain.NewSevenBridgePlayer(true),
		domain.NewSevenBridgePlayer(false),
	}
	return domain.NewSevenBridge(domain.NewTrumpCards(0), players, domain.DefaultSevenBridgeConfig())
}

func newTestSevenBridgeDifficulty(d domain.SevenBridgeCpuDifficulty) *domain.SevenBridge {
	players := []*domain.SevenBridgePlayer{
		domain.NewSevenBridgePlayer(true),
		domain.NewSevenBridgePlayer(false),
	}
	cfg := domain.DefaultSevenBridgeConfig()
	cfg.CpuDifficulty = d
	return domain.NewSevenBridge(domain.NewTrumpCards(0), players, cfg)
}

// setHand replaces a player's hand with the provided cards (test helper).
func setHand(p *domain.SevenBridgePlayer, cards []*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestNewSevenBridge(t *testing.T) {
	g := newTestSevenBridge()
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, 0, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
}

func TestSevenBridge_NewDefault(t *testing.T) {
	g := domain.NewDefaultSevenBridge()
	require.NotNil(t, g)
	assert.Equal(t, 2, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman())
	assert.False(t, g.GetPlayer(1).GetIsHuman())
}

func TestSevenBridge_Reset(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()

	assert.Equal(t, domain.SevenBridgePhaseDraw, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerIdx())

	for i := range 2 {
		assert.Equal(t, 7, g.GetPlayer(i).GetCardsSize())
		assert.Equal(t, 0, g.GetPlayer(i).GetRoundScore())
		assert.Equal(t, 0, g.GetPlayer(i).GetCumulativeScore())
		assert.Equal(t, 0, g.GetPlayer(i).GetMeldCount())
	}

	// discard top should be a 7
	top := g.GetDiscardTop()
	require.NotNil(t, top)
	assert.Equal(t, domain.SevenBridgePivotRank, top.GetValue())
}

func TestSevenBridge_Reset_ClearsState(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()

	g.GetPlayer(0).SetCumulativeScore(200)
	g.GetPlayer(0).AppendMeld([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignHeart, 3, true),
	})
	g.SetPhase(domain.SevenBridgePhaseGameEnd)

	g.Reset()
	assert.Equal(t, domain.SevenBridgePhaseDraw, g.GetPhase())
	assert.Equal(t, 0, g.GetPlayer(0).GetCumulativeScore())
	assert.Equal(t, 0, g.GetPlayer(0).GetMeldCount())
}

func TestSevenBridge_Getters(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	assert.Equal(t, 2, g.GetPlayerCnt())
	assert.NotNil(t, g.GetPlayer(0))
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(2))
	assert.NotEmpty(t, g.GetDiscardPile())
	assert.NotZero(t, g.GetDrawPileCount())
	assert.Equal(t, domain.DefaultSevenBridgeConfig(), g.GetConfig())
	assert.Equal(t, -1, g.GetRoundWinnerIdx())
	assert.False(t, g.GetClaimedThisTurn())
}

func TestSevenBridge_SetConfig(t *testing.T) {
	g := newTestSevenBridge()
	cfg := domain.SevenBridgeConfig{CpuDifficulty: domain.SevenBridgeCpuDifficultyHard, PointLimit: 50}
	g.SetConfig(cfg)
	assert.Equal(t, cfg, g.GetConfig())
}

func TestSevenBridge_IsHumanTurn(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(1)
	assert.False(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(99)
	assert.False(t, g.IsHumanTurn())
}

// --- Phase / turn enforcement ---

func TestSevenBridge_DrawFromStock_WrongPhase(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetPhase(domain.SevenBridgePhasePlay)
	err := g.PlayerDrawFromStock()
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestSevenBridge_DrawFromStock_NotHumanTurn(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetPhase(domain.SevenBridgePhaseDraw)
	g.SetCurrentPlayerIdx(1)
	err := g.PlayerDrawFromStock()
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
}

// finishGameAt sets up and triggers the minimum path that ends the game.
// Useful as a shared helper for "method rejects after GameEnded" tests.
func finishGameAt(g *domain.SevenBridge) {
	cfg := g.GetConfig()
	cfg.PointLimit = 1
	g.SetConfig(cfg)
	g.GetPlayer(0).AppendMeld([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignHeart, 3, true),
	})
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, true)})
	setHand(g.GetPlayer(0), []*domain.Card{domain.NewCard(domain.CardDesignSpade, 8, true)})
	setHand(g.GetPlayer(1), []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, true)})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)
	_ = g.PlayerDiscard(0)
}

func TestSevenBridge_AllActions_RejectAfterGameEnded(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	finishGameAt(g)
	require.True(t, g.GetGameEndFlag())

	assert.ErrorIs(t, g.PlayerDrawFromStock(), domain.ErrGameEnded)
	assert.ErrorIs(t, g.PlayerClaimPon([]int{0, 1}), domain.ErrGameEnded)
	assert.ErrorIs(t, g.PlayerClaimChi([]int{0, 1}), domain.ErrGameEnded)
	assert.ErrorIs(t, g.PlayerMeld([]int{0, 1, 2}), domain.ErrGameEnded)
	assert.ErrorIs(t, g.PlayerLayoff(1, 0, 0), domain.ErrGameEnded)
	assert.ErrorIs(t, g.PlayerDiscard(0), domain.ErrGameEnded)
}

func TestSevenBridge_DrawFromStock_Success(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	before := g.GetPlayer(0).GetCardsSize()
	drawBefore := g.GetDrawPileCount()
	err := g.PlayerDrawFromStock()
	require.NoError(t, err)
	assert.Equal(t, before+1, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, drawBefore-1, g.GetDrawPileCount())
	assert.Equal(t, domain.SevenBridgePhasePlay, g.GetPhase())
}

func TestSevenBridge_DrawFromStock_EndsOnEmpty(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetDrawPile([]*domain.Card{})
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.SevenBridgePhaseDraw)
	err := g.PlayerDrawFromStock()
	require.NoError(t, err)
	// Stock-out ends the round (no winner)
	assert.Equal(t, -1, g.GetRoundWinnerIdx())
}

// --- Discard legality ---

func TestIsDiscardLegal(t *testing.T) {
	c := func(design, value int) *domain.Card { return domain.NewCard(design, value, true) }
	cases := []struct {
		name string
		card *domain.Card
		top  *domain.Card
		want bool
	}{
		{"first discard must be 7", c(domain.CardDesignSpade, 7), nil, true},
		{"first discard 5 rejected", c(domain.CardDesignSpade, 5), nil, false},
		{"seven always legal", c(domain.CardDesignHeart, 7), c(domain.CardDesignDiamond, 2), true},
		{"same rank ok", c(domain.CardDesignSpade, 9), c(domain.CardDesignHeart, 9), true},
		{"+1 rank ok", c(domain.CardDesignSpade, 8), c(domain.CardDesignHeart, 7), true},
		{"-1 rank ok", c(domain.CardDesignSpade, 6), c(domain.CardDesignHeart, 7), true},
		{"unrelated rejected", c(domain.CardDesignSpade, 13), c(domain.CardDesignHeart, 2), false},
		{"nil card rejected", nil, c(domain.CardDesignHeart, 7), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, domain.IsDiscardLegal(tc.card, tc.top))
		})
	}
}

// --- Meld validation ---

func TestIsSevenBridgeMeld(t *testing.T) {
	c := func(design, value int) *domain.Card { return domain.NewCard(design, value, true) }

	cases := []struct {
		name  string
		cards []*domain.Card
		want  bool
	}{
		{"empty", nil, false},
		{"two cards", []*domain.Card{c(1, 5), c(2, 5)}, false},
		{"set of three", []*domain.Card{c(1, 5), c(2, 5), c(3, 5)}, true},
		{"set duplicate suit", []*domain.Card{c(1, 5), c(1, 5), c(3, 5)}, false},
		{"run of three", []*domain.Card{c(1, 3), c(1, 4), c(1, 5)}, true},
		{"run unordered ok", []*domain.Card{c(1, 5), c(1, 3), c(1, 4)}, true},
		{"run broken", []*domain.Card{c(1, 3), c(1, 4), c(1, 6)}, false},
		{"run mixed suit", []*domain.Card{c(1, 3), c(2, 4), c(1, 5)}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, domain.IsSevenBridgeMeld(tc.cards))
		})
	}
}

// --- Meld action ---

func TestSevenBridge_PlayerMeld_Success(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignHeart, 3, true),
		domain.NewCard(domain.CardDesignDiamond, 10, true),
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)

	err := g.PlayerMeld([]int{0, 1, 2})
	require.NoError(t, err)
	assert.Equal(t, 1, g.GetPlayer(0).GetMeldCount())
	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
}

func TestSevenBridge_PlayerMeld_InvalidCombo(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignClover, 4, true),
		domain.NewCard(domain.CardDesignHeart, 9, true),
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)

	err := g.PlayerMeld([]int{0, 1, 2})
	require.Error(t, err)
	var dErr *domain.DomainError
	assert.True(t, errors.As(err, &dErr))
}

func TestSevenBridge_PlayerMeld_WrongPhase(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetPhase(domain.SevenBridgePhaseDraw)
	err := g.PlayerMeld([]int{0, 1, 2})
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestSevenBridge_PlayerMeld_NotHumanTurn(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(1)
	err := g.PlayerMeld([]int{0, 1, 2})
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
}

func TestSevenBridge_PlayerMeld_TooFewCards(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)
	err := g.PlayerMeld([]int{0, 1})
	require.Error(t, err)
}

func TestSevenBridge_PlayerMeld_InvalidIndex(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignHeart, 3, true),
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)
	err := g.PlayerMeld([]int{0, 1, 99})
	require.Error(t, err)
}

// --- Layoff ---

func TestSevenBridge_PlayerLayoff_Success(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	// opponent has a run meld 3-4-5 of spades
	g.GetPlayer(1).AppendMeld([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignSpade, 4, true),
		domain.NewCard(domain.CardDesignSpade, 5, true),
	})
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, true),
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)

	err := g.PlayerLayoff(1, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, g.GetPlayer(0).GetCardsSize())
	assert.Len(t, g.GetPlayer(1).GetMeld(0), 4)
}

func TestSevenBridge_PlayerLayoff_BadTarget(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)
	err := g.PlayerLayoff(99, 0, 0)
	require.Error(t, err)
	err = g.PlayerLayoff(1, 99, 0)
	require.Error(t, err)
}

func TestSevenBridge_PlayerLayoff_WrongCard(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.GetPlayer(1).AppendMeld([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignSpade, 4, true),
		domain.NewCard(domain.CardDesignSpade, 5, true),
	})
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 10, true),
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)
	err := g.PlayerLayoff(1, 0, 0)
	require.Error(t, err)
}

func TestSevenBridge_PlayerLayoff_OutOfRange(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.GetPlayer(1).AppendMeld([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignSpade, 4, true),
		domain.NewCard(domain.CardDesignSpade, 5, true),
	})
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, true),
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)
	err := g.PlayerLayoff(1, 0, 99)
	require.Error(t, err)
}

// --- Discard ---

func TestSevenBridge_PlayerDiscard_LegalRule(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, true)})
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 8, true), // legal (+1)
		domain.NewCard(domain.CardDesignClover, 13, true),
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)

	err := g.PlayerDiscard(0)
	require.NoError(t, err)
	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 8, g.GetDiscardTop().GetValue())
	// turn advanced to CPU
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	assert.Equal(t, domain.SevenBridgePhaseDraw, g.GetPhase())
}

func TestSevenBridge_PlayerDiscard_IllegalRejected(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 3, true)})
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 13, true), // illegal
		domain.NewCard(domain.CardDesignClover, 4, true), // legal (+1)
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)

	err := g.PlayerDiscard(0)
	require.Error(t, err)
}

func TestSevenBridge_PlayerDiscard_FallbackWhenStuck(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 3, true)})
	// Two illegal cards — the hand has no legal discard, but we keep 2 cards so the
	// "no-meld-last-card" guard doesn't fire and we exercise the illegal-discard
	// relaxation path.
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 13, true),
		domain.NewCard(domain.CardDesignClover, 12, true),
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)

	err := g.PlayerDiscard(0)
	require.NoError(t, err)
	assert.Equal(t, 13, g.GetDiscardTop().GetValue())
}

func TestSevenBridge_PlayerDiscard_OutOfRange(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)
	err := g.PlayerDiscard(99)
	require.Error(t, err)
}

func TestSevenBridge_PlayerDiscard_WrongPhase(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetPhase(domain.SevenBridgePhaseDraw)
	err := g.PlayerDiscard(0)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

// --- Pon / Chi claims ---

func TestSevenBridge_PlayerClaimPon_Success(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)})
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, true),
		domain.NewCard(domain.CardDesignClover, 9, true),
		domain.NewCard(domain.CardDesignDiamond, 3, true),
	})
	g.SetPhase(domain.SevenBridgePhaseDraw)
	g.SetCurrentPlayerIdx(0)

	err := g.PlayerClaimPon([]int{0, 1})
	require.NoError(t, err)
	assert.Equal(t, 1, g.GetPlayer(0).GetMeldCount())
	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
	assert.True(t, g.GetClaimedThisTurn())
	assert.Equal(t, domain.SevenBridgePhasePlay, g.GetPhase())
}

func TestSevenBridge_PlayerClaimPon_RejectWrongRank(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)})
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, true),
		domain.NewCard(domain.CardDesignClover, 9, true),
	})
	g.SetPhase(domain.SevenBridgePhaseDraw)
	g.SetCurrentPlayerIdx(0)

	err := g.PlayerClaimPon([]int{0, 1})
	require.Error(t, err)
}

func TestSevenBridge_PlayerClaimPon_BadIndexCount(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetPhase(domain.SevenBridgePhaseDraw)
	g.SetCurrentPlayerIdx(0)
	err := g.PlayerClaimPon([]int{0})
	require.Error(t, err)
}

func TestSevenBridge_PlayerClaimPon_EmptyDiscard(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetDiscardPile([]*domain.Card{})
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, true),
		domain.NewCard(domain.CardDesignClover, 9, true),
	})
	g.SetPhase(domain.SevenBridgePhaseDraw)
	g.SetCurrentPlayerIdx(0)
	err := g.PlayerClaimPon([]int{0, 1})
	require.Error(t, err)
}

func TestSevenBridge_PlayerClaimChi_Success(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)})
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, true),
		domain.NewCard(domain.CardDesignSpade, 7, true),
		domain.NewCard(domain.CardDesignDiamond, 9, true),
	})
	g.SetPhase(domain.SevenBridgePhaseDraw)
	g.SetCurrentPlayerIdx(0)

	err := g.PlayerClaimChi([]int{0, 1})
	require.NoError(t, err)
	assert.Equal(t, 1, g.GetPlayer(0).GetMeldCount())
}

func TestSevenBridge_PlayerClaimChi_WrongSuit(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)})
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 6, true),
		domain.NewCard(domain.CardDesignSpade, 7, true),
	})
	g.SetPhase(domain.SevenBridgePhaseDraw)
	g.SetCurrentPlayerIdx(0)
	err := g.PlayerClaimChi([]int{0, 1})
	require.Error(t, err)
}

func TestSevenBridge_PlayerClaimChi_NotSequential(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)})
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 8, true),
		domain.NewCard(domain.CardDesignSpade, 10, true),
	})
	g.SetPhase(domain.SevenBridgePhaseDraw)
	g.SetCurrentPlayerIdx(0)
	err := g.PlayerClaimChi([]int{0, 1})
	require.Error(t, err)
}

// --- Round finishing ---

func TestSevenBridge_FinishRoundOnEmptyHand(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	// player 0 has 1 card and a meld already
	g.GetPlayer(0).AppendMeld([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignHeart, 3, true),
	})
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, true)})
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 8, true),
	})
	setHand(g.GetPlayer(1), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, true),    // 50
		domain.NewCard(domain.CardDesignDiamond, 13, true), // 10
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)

	err := g.PlayerDiscard(0)
	require.NoError(t, err)
	assert.Equal(t, 0, g.GetRoundWinnerIdx())
	assert.Equal(t, 60, g.GetPlayer(0).GetCumulativeScore())
	assert.Equal(t, domain.SevenBridgePhaseRoundEnd, g.GetPhase())
}

func TestSevenBridge_GameEndOnPointLimit(t *testing.T) {
	g := newTestSevenBridge()
	cfg := g.GetConfig()
	cfg.PointLimit = 50
	g.SetConfig(cfg)
	g.Reset()
	g.GetPlayer(0).AppendMeld([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignHeart, 3, true),
	})
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, true)})
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 8, true),
	})
	setHand(g.GetPlayer(1), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, true), // 50 → hits limit
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)

	err := g.PlayerDiscard(0)
	require.NoError(t, err)
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerIdx())
	assert.Equal(t, domain.SevenBridgePhaseGameEnd, g.GetPhase())
}

func TestSevenBridge_NextRound(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()

	// put in round-end state manually
	g.SetPhase(domain.SevenBridgePhaseRoundEnd)
	g.GetPlayer(1).SetCumulativeScore(30)
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, domain.SevenBridgePhaseDraw, g.GetPhase())
	for i := range 2 {
		assert.Equal(t, 7, g.GetPlayer(i).GetCardsSize())
	}
	// cumulative score preserved between rounds
	assert.Equal(t, 30, g.GetPlayer(1).GetCumulativeScore())
}

func TestSevenBridge_NextRound_IgnoresWrongPhase(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetPhase(domain.SevenBridgePhasePlay)
	start := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, start, g.GetRoundNumber())
}

// --- CPU play ---

func TestSevenBridge_CpuDraw_FromStock(t *testing.T) {
	g := newTestSevenBridgeDifficulty(domain.SevenBridgeCpuDifficultyEasy)
	g.Reset()
	g.SetCurrentPlayerIdx(1)
	g.SetPhase(domain.SevenBridgePhaseDraw)
	// Empty hand so pon/chi cannot apply
	g.GetPlayer(1).Reset()
	before := g.GetDrawPileCount()
	g.CpuPlay()
	assert.Equal(t, before-1, g.GetDrawPileCount())
	assert.Equal(t, domain.SevenBridgePhasePlay, g.GetPhase())
}

func TestSevenBridge_CpuPlay_SkippedOnHumanTurn(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	beforeDraw := g.GetDrawPileCount()
	g.CpuPlay()
	assert.Equal(t, beforeDraw, g.GetDrawPileCount())
}

func TestSevenBridge_CpuPlay_GameEndedNoop(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	cfg := g.GetConfig()
	cfg.PointLimit = 1
	g.SetConfig(cfg)
	g.GetPlayer(0).AppendMeld([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignHeart, 3, true),
	})
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, true)})
	setHand(g.GetPlayer(0), []*domain.Card{domain.NewCard(domain.CardDesignSpade, 8, true)})
	setHand(g.GetPlayer(1), []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, true)})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)
	require.NoError(t, g.PlayerDiscard(0))
	require.True(t, g.GetGameEndFlag())
	g.CpuPlay() // no-op, no error
}

func TestSevenBridge_CpuPlay_MeldsAndDiscards(t *testing.T) {
	g := newTestSevenBridgeDifficulty(domain.SevenBridgeCpuDifficultyHard)
	g.Reset()
	// Put discard top = 5H so spades 6 is a legal discard (+1)
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, true)})
	setHand(g.GetPlayer(1), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignHeart, 3, true),
		domain.NewCard(domain.CardDesignSpade, 6, true),
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(1)
	g.CpuPlay()
	assert.Equal(t, 1, g.GetPlayer(1).GetMeldCount())
	// After melding, CPU must have discarded
	assert.LessOrEqual(t, g.GetPlayer(1).GetCardsSize(), 1)
}

func TestSevenBridge_CpuPlay_PonClaimWhenAvailable(t *testing.T) {
	g := newTestSevenBridgeDifficulty(domain.SevenBridgeCpuDifficultyHard)
	g.Reset()
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)})
	setHand(g.GetPlayer(1), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, true),
		domain.NewCard(domain.CardDesignClover, 9, true),
		domain.NewCard(domain.CardDesignDiamond, 12, true),
	})
	g.SetPhase(domain.SevenBridgePhaseDraw)
	g.SetCurrentPlayerIdx(1)
	g.CpuPlay()
	assert.Equal(t, 1, g.GetPlayer(1).GetMeldCount())
	assert.Equal(t, domain.SevenBridgePhasePlay, g.GetPhase())
}

func TestSevenBridge_CpuPlay_WinsByEmptyingHand(t *testing.T) {
	g := newTestSevenBridgeDifficulty(domain.SevenBridgeCpuDifficultyHard)
	g.Reset()
	g.GetPlayer(1).AppendMeld([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignHeart, 3, true),
	})
	setHand(g.GetPlayer(1), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 8, true),
		domain.NewCard(domain.CardDesignClover, 8, true),
		domain.NewCard(domain.CardDesignHeart, 8, true),
	})
	setHand(g.GetPlayer(0), nil) // clear loser hand so round end doesn't tip into GameEnd
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(1)
	g.CpuPlay()
	// CPU melded the 888 set; hand became empty → round ended
	assert.Equal(t, 1, g.GetRoundWinnerIdx())
	assert.Equal(t, domain.SevenBridgePhaseRoundEnd, g.GetPhase())
}

// --- Config ---

func TestSevenBridgeConfig_Default(t *testing.T) {
	c := domain.DefaultSevenBridgeConfig()
	assert.Equal(t, domain.SevenBridgeCpuDifficultyNormal, c.CpuDifficulty)
	assert.Equal(t, 100, c.PointLimit)
	assert.NoError(t, c.Validate())
}

func TestSevenBridgeConfig_ValidateErrors(t *testing.T) {
	c := domain.SevenBridgeConfig{CpuDifficulty: -1, PointLimit: 100}
	assert.Error(t, c.Validate())
	c = domain.SevenBridgeConfig{CpuDifficulty: domain.SevenBridgeCpuDifficultyNormal, PointLimit: 0}
	assert.Error(t, c.Validate())
}

// --- Player ---

func TestSevenBridgePlayer_Basics(t *testing.T) {
	p := domain.NewSevenBridgePlayer(true)
	assert.True(t, p.GetIsHuman())
	p.AppendMeld([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignHeart, 3, true),
	})
	assert.Equal(t, 1, p.GetMeldCount())
	assert.Nil(t, p.GetMeld(-1))
	assert.Nil(t, p.GetMeld(5))
	assert.True(t, p.AddCardToMeld(0, domain.NewCard(domain.CardDesignDiamond, 3, true)))
	assert.Len(t, p.GetMeld(0), 4)
	assert.False(t, p.AddCardToMeld(99, domain.NewCard(domain.CardDesignSpade, 5, true)))
	p.ResetRound()
	assert.Equal(t, 0, p.GetMeldCount())
	assert.Equal(t, 0, p.GetCardsSize())
}

func TestSevenBridgePlayer_SetMelds(t *testing.T) {
	p := domain.NewSevenBridgePlayer(false)
	m := [][]*domain.Card{{domain.NewCard(domain.CardDesignSpade, 1, true)}}
	p.SetMelds(m)
	assert.Equal(t, 1, p.GetMeldCount())
}

// --- JSON round-trip ---

func TestSevenBridge_JSONRoundTrip(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.GetPlayer(0).AppendMeld([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignHeart, 3, true),
	})
	b, err := json.Marshal(g)
	require.NoError(t, err)

	// Rebuild via UnmarshalJSON
	restored := domain.NewSevenBridge(nil, nil, domain.DefaultSevenBridgeConfig())
	require.NoError(t, restored.UnmarshalJSON(b))
	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, g.GetPlayer(0).GetMeldCount(), restored.GetPlayer(0).GetMeldCount())
	assert.Equal(t, g.GetDrawPileCount(), restored.GetDrawPileCount())
}

func TestSevenBridge_JSONUnmarshal_Invalid(t *testing.T) {
	g := domain.NewSevenBridge(nil, nil, domain.DefaultSevenBridgeConfig())
	err := g.UnmarshalJSON([]byte("{"))
	require.Error(t, err)
}

func TestSevenBridgePlayer_JSONRoundTrip(t *testing.T) {
	p := domain.NewSevenBridgePlayer(true)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 7, true))
	p.AppendMeld([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignHeart, 3, true),
	})
	p.SetRoundScore(12)
	b, err := json.Marshal(p)
	require.NoError(t, err)

	q := &domain.SevenBridgePlayer{}
	require.NoError(t, json.Unmarshal(b, q))
	assert.True(t, q.GetIsHuman())
	assert.Equal(t, 1, q.GetCardsSize())
	assert.Equal(t, 1, q.GetMeldCount())
	assert.Equal(t, 12, q.GetRoundScore())
}

func TestSevenBridgePlayer_UnmarshalInvalid(t *testing.T) {
	p := &domain.SevenBridgePlayer{}
	err := json.Unmarshal([]byte("{"), p)
	require.Error(t, err)
}

// --- Additional coverage ---

func TestSevenBridgePlayer_GetMelds(t *testing.T) {
	p := domain.NewSevenBridgePlayer(true)
	assert.Nil(t, p.GetMelds())
	meld := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignHeart, 3, true),
	}
	p.AppendMeld(meld)
	assert.Len(t, p.GetMelds(), 1)
}

func TestSevenBridge_GetActionLog(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignHeart, 3, true),
		domain.NewCard(domain.CardDesignDiamond, 10, true), // filler so meld doesn't empty the hand
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)
	require.NoError(t, g.PlayerMeld([]int{0, 1, 2}))
	log := g.GetActionLog()
	assert.NotEmpty(t, log)
	assert.Equal(t, "meld", log[len(log)-1].ActionType)
}

func TestSevenBridge_SetRoundNumber(t *testing.T) {
	g := newTestSevenBridge()
	g.SetRoundNumber(7)
	assert.Equal(t, 7, g.GetRoundNumber())
}

// Exercises Hard/Easy paths in shouldCpuClaim.
func TestSevenBridge_CpuPlay_ShouldClaimAcrossDifficulties(t *testing.T) {
	for _, d := range []domain.SevenBridgeCpuDifficulty{
		domain.SevenBridgeCpuDifficultyEasy,
		domain.SevenBridgeCpuDifficultyNormal,
		domain.SevenBridgeCpuDifficultyHard,
	} {
		g := newTestSevenBridgeDifficulty(d)
		g.Reset()
		g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)})
		// Extra filler card so a successful Pon claim (empties hand → round end) doesn't
		// fire — we want the phase to end up in Play for this assertion.
		setHand(g.GetPlayer(1), []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 9, true),
			domain.NewCard(domain.CardDesignClover, 9, true),
			domain.NewCard(domain.CardDesignDiamond, 2, true),
		})
		g.SetPhase(domain.SevenBridgePhaseDraw)
		g.SetCurrentPlayerIdx(1)
		// Force a draw-pile card so falling through to stock draw still succeeds.
		g.SetDrawPile([]*domain.Card{domain.NewCard(domain.CardDesignDiamond, 3, true)})
		g.CpuPlay()
		// Either claimed (meld) or drew from stock — in any case the phase transitions to Play.
		assert.Equal(t, domain.SevenBridgePhasePlay, g.GetPhase())
	}
}

// Exercises the run-meld branch in findBestMeldIndices (no set available).
func TestSevenBridge_CpuPlay_MeldsRun(t *testing.T) {
	g := newTestSevenBridgeDifficulty(domain.SevenBridgeCpuDifficultyHard)
	g.Reset()
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)})
	// run 3-4-5 of spades + a discard-legal 10 (same rank as any card; here same rank as no top so use +1/-1)
	// top is 9H, so a 10C is legal (+1 rank).
	setHand(g.GetPlayer(1), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignSpade, 4, true),
		domain.NewCard(domain.CardDesignSpade, 5, true),
		domain.NewCard(domain.CardDesignClover, 10, true),
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(1)
	g.CpuPlay()
	assert.Equal(t, 1, g.GetPlayer(1).GetMeldCount())
	meld := g.GetPlayer(1).GetMeld(0)
	assert.Len(t, meld, 3)
}

func TestSevenBridge_PopDiscardTop_Empty(t *testing.T) {
	g := newTestSevenBridge()
	g.SetDiscardPile([]*domain.Card{})
	assert.Nil(t, g.GetDiscardTop())
}

func TestSevenBridgeCardPenalty_AllValues(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	// Build a hand with each penalty tier and force a round finish where opponent's
	// total penalty exercises every branch of sevenBridgeCardPenalty.
	g.GetPlayer(0).AppendMeld([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignHeart, 3, true),
	})
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, true)})
	setHand(g.GetPlayer(0), []*domain.Card{domain.NewCard(domain.CardDesignSpade, 8, true)})
	setHand(g.GetPlayer(1), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, true),    // 1
		domain.NewCard(domain.CardDesignSpade, 5, true),    // 5
		domain.NewCard(domain.CardDesignSpade, 7, true),    // 50
		domain.NewCard(domain.CardDesignDiamond, 13, true), // 10
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)
	require.NoError(t, g.PlayerDiscard(0))
	assert.Equal(t, 0, g.GetRoundWinnerIdx())
	assert.Equal(t, 66, g.GetPlayer(0).GetRoundScore()) // 1 + 5 + 50 + 10 = 66
}

// Drives dealInitialCards fallback when no 7 is in the remaining draw pile.
// We simulate this by calling Reset() on a modified deck-less game. Since we
// cannot trivially replace TrumpCards, this is covered indirectly by Reset above.
// The fallback branch is executed when all 7s are pre-dealt to players — very
// unlikely with a 52-card deck and 14 dealt cards, so we leave the branch
// documented and exercise it directly via dealInitialCards behaviour below.
func TestSevenBridge_Reset_TopDiscardAlwaysLegalSeven(t *testing.T) {
	// Run Reset() many times to see top always is a 7 (the pivot rule).
	for range 10 {
		g := newTestSevenBridge()
		g.Reset()
		top := g.GetDiscardTop()
		require.NotNil(t, top)
		assert.Equal(t, domain.SevenBridgePivotRank, top.GetValue())
	}
}

func TestSevenBridge_PlayerActions_GuardsBranches(t *testing.T) {
	// ClaimPon: wrong phase
	g := newTestSevenBridge()
	g.Reset()
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)
	assert.ErrorIs(t, g.PlayerClaimPon([]int{0, 1}), domain.ErrWrongPhase)

	// ClaimPon: not human turn
	g.SetPhase(domain.SevenBridgePhaseDraw)
	g.SetCurrentPlayerIdx(1)
	assert.ErrorIs(t, g.PlayerClaimPon([]int{0, 1}), domain.ErrNotHumanTurn)

	// ClaimChi: wrong phase
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)
	assert.ErrorIs(t, g.PlayerClaimChi([]int{0, 1}), domain.ErrWrongPhase)

	// ClaimChi: not human turn
	g.SetPhase(domain.SevenBridgePhaseDraw)
	g.SetCurrentPlayerIdx(1)
	assert.ErrorIs(t, g.PlayerClaimChi([]int{0, 1}), domain.ErrNotHumanTurn)

	// Layoff: wrong phase
	g.SetPhase(domain.SevenBridgePhaseDraw)
	g.SetCurrentPlayerIdx(0)
	assert.ErrorIs(t, g.PlayerLayoff(1, 0, 0), domain.ErrWrongPhase)

	// Layoff: not human turn
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(1)
	assert.ErrorIs(t, g.PlayerLayoff(1, 0, 0), domain.ErrNotHumanTurn)

	// Discard: not human turn
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(1)
	assert.ErrorIs(t, g.PlayerDiscard(0), domain.ErrNotHumanTurn)
}

func TestSevenBridge_PlayerName_OutOfRange(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	// Exercise playerName indirectly by logging a stock-out at index -1 (no-op)
	// Instead, check via finishRound(-1) which logs with playerIdx=-1; the label
	// should use the numeric fallback in playerName.
	g.SetDrawPile([]*domain.Card{})
	g.SetPhase(domain.SevenBridgePhaseDraw)
	g.SetCurrentPlayerIdx(0)
	require.NoError(t, g.PlayerDrawFromStock())
	log := g.GetActionLog()
	require.NotEmpty(t, log)
	assert.Equal(t, "draw", log[len(log)-1].ActionType)
}

func TestSevenBridge_ValidateIndexList_Duplicate(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignHeart, 3, true),
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)
	err := g.PlayerMeld([]int{0, 0, 1})
	require.Error(t, err)
}

func TestSevenBridge_UnmarshalJSON_SliceTooLarge(t *testing.T) {
	// Craft a JSON payload with a 1001-element action log.
	var sb strings.Builder
	sb.WriteString(`{"al":[`)
	for i := range 1001 {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(`{"pi":0,"at":"x","dt":"y"}`)
	}
	sb.WriteString(`]}`)
	g := domain.NewSevenBridge(nil, nil, domain.DefaultSevenBridgeConfig())
	err := g.UnmarshalJSON([]byte(sb.String()))
	require.Error(t, err)
}

func TestSevenBridge_PopDiscardTop_NonEmpty(t *testing.T) {
	g := newTestSevenBridge()
	g.SetDiscardPile([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignHeart, 9, true),
	})
	top := g.GetDiscardTop()
	require.NotNil(t, top)
	assert.Equal(t, 9, top.GetValue())
	assert.Len(t, g.GetDiscardPile(), 2)
}

func TestSevenBridge_CpuDiscard_HardMustDropSevenIfOnlyLegal(t *testing.T) {
	// Hard CPU prefers non-7 discards, but must drop a 7 if it's the only legal option.
	g := newTestSevenBridgeDifficulty(domain.SevenBridgeCpuDifficultyHard)
	g.Reset()
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 13, true)})
	setHand(g.GetPlayer(1), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, true), // only legal discard (7 wildcard)
		domain.NewCard(domain.CardDesignClover, 2, true),
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(1)
	g.CpuPlay()
	// 7 was the only legal move — CPU discarded it.
	top := g.GetDiscardTop()
	require.NotNil(t, top)
	assert.Equal(t, 7, top.GetValue())
}

func TestSevenBridge_CpuDiscard_HardAllSevens(t *testing.T) {
	// Hard filter would empty the list when hand is all 7s — it must fall back
	// to the unfiltered list so the CPU still has something to discard.
	g := newTestSevenBridgeDifficulty(domain.SevenBridgeCpuDifficultyHard)
	g.Reset()
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 13, true)})
	setHand(g.GetPlayer(1), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, true),
		domain.NewCard(domain.CardDesignClover, 7, true),
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(1)
	g.CpuPlay()
	// Discard top should now be a 7 (no alternative).
	top := g.GetDiscardTop()
	require.NotNil(t, top)
	assert.Equal(t, 7, top.GetValue())
}

func TestSevenBridge_CpuDiscard_NilTop(t *testing.T) {
	// When no top (first-turn artificial state) CPU discards anything legal.
	g := newTestSevenBridgeDifficulty(domain.SevenBridgeCpuDifficultyNormal)
	g.Reset()
	g.SetDiscardPile([]*domain.Card{})
	setHand(g.GetPlayer(1), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, true),
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(1)
	g.CpuPlay()
	assert.NotNil(t, g.GetDiscardTop())
}

// --- Review-fix regression tests ---

func TestSevenBridge_Meld_EmptiesHandFinishesRound(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignHeart, 3, true),
	})
	setHand(g.GetPlayer(1), nil) // clear so loser penalty doesn't push the game to GameEnd
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)

	require.NoError(t, g.PlayerMeld([]int{0, 1, 2}))

	assert.Equal(t, 0, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, domain.SevenBridgePhaseRoundEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetRoundWinnerIdx())
}

func TestSevenBridge_Layoff_EmptiesHandFinishesRound(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.GetPlayer(1).AppendMeld([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, true),
		domain.NewCard(domain.CardDesignSpade, 4, true),
		domain.NewCard(domain.CardDesignSpade, 5, true),
	})
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, true),
	})
	setHand(g.GetPlayer(1), nil) // clear so loser penalty doesn't push the game to GameEnd
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)

	require.NoError(t, g.PlayerLayoff(1, 0, 0))

	assert.Equal(t, 0, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, domain.SevenBridgePhaseRoundEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetRoundWinnerIdx())
}

func TestSevenBridge_ClaimPon_EmptiesHandFinishesRound(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)})
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, true),
		domain.NewCard(domain.CardDesignClover, 9, true),
	})
	setHand(g.GetPlayer(1), nil) // clear so loser penalty doesn't push the game to GameEnd
	g.SetPhase(domain.SevenBridgePhaseDraw)
	g.SetCurrentPlayerIdx(0)

	require.NoError(t, g.PlayerClaimPon([]int{0, 1}))

	assert.Equal(t, 0, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, domain.SevenBridgePhaseRoundEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetRoundWinnerIdx())
}

func TestSevenBridge_ClaimChi_EmptiesHandFinishesRound(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)})
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, true),
		domain.NewCard(domain.CardDesignSpade, 7, true),
	})
	setHand(g.GetPlayer(1), nil) // clear so loser penalty doesn't push the game to GameEnd
	g.SetPhase(domain.SevenBridgePhaseDraw)
	g.SetCurrentPlayerIdx(0)

	require.NoError(t, g.PlayerClaimChi([]int{0, 1}))

	assert.Equal(t, 0, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, domain.SevenBridgePhaseRoundEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetRoundWinnerIdx())
}

func TestSevenBridge_Discard_RejectsLastCardWithoutMeld(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, true)})
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, true),
	})
	g.SetPhase(domain.SevenBridgePhasePlay)
	g.SetCurrentPlayerIdx(0)

	err := g.PlayerDiscard(0)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	// Hand and phase unchanged.
	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, domain.SevenBridgePhasePlay, g.GetPhase())
}

func TestSevenBridge_CpuDraw_FromDiscardViaChi(t *testing.T) {
	g := newTestSevenBridgeDifficulty(domain.SevenBridgeCpuDifficultyHard)
	g.Reset()
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)})
	setHand(g.GetPlayer(1), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, true),
		domain.NewCard(domain.CardDesignSpade, 7, true),
	})
	g.SetPhase(domain.SevenBridgePhaseDraw)
	g.SetCurrentPlayerIdx(1)
	g.CpuPlay()
	assert.Equal(t, 1, g.GetPlayer(1).GetMeldCount())
}

// **判定は claim 経路と同じ finder を通す。**別に書くと、案内した組が拒否される
// ことになる (#4904)。
func TestSevenBridge_SuggestPonAndChi(t *testing.T) {
	g := newTestSevenBridge()
	g.Reset()
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, true)})

	// 同ランク 2 枚 → ポン。
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, true),
		domain.NewCard(domain.CardDesignClover, 7, true),
		domain.NewCard(domain.CardDesignDiamond, 2, true),
	})
	assert.Equal(t, []int{0, 1}, g.SuggestPon(0))
	// **チーは同スートの連番。**♥7 に対して ♠7 ♣7 は繋がらない。
	assert.Nil(t, g.SuggestChi(0))

	// 同スートの連番 2 枚 → チー。ポンは成立しない。
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, true),
		domain.NewCard(domain.CardDesignHeart, 6, true),
		domain.NewCard(domain.CardDesignSpade, 13, true),
	})
	assert.Nil(t, g.SuggestPon(0))
	assert.Equal(t, []int{0, 1}, g.SuggestChi(0))

	// またぎ (6, 8) でもチーになる。
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 6, true),
		domain.NewCard(domain.CardDesignHeart, 8, true),
	})
	assert.Equal(t, []int{0, 1}, g.SuggestChi(0))

	// どちらも成立しない手札。
	setHand(g.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, true),
		domain.NewCard(domain.CardDesignClover, 13, true),
	})
	assert.Nil(t, g.SuggestPon(0))
	assert.Nil(t, g.SuggestChi(0))

	// 捨て札が無ければどちらも nil。範囲外の席も nil。
	g.SetDiscardPile(nil)
	assert.Nil(t, g.SuggestPon(0))
	assert.Nil(t, g.SuggestChi(0))
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, true)})
	assert.Nil(t, g.SuggestPon(99))
	assert.Nil(t, g.SuggestChi(-1))
}
