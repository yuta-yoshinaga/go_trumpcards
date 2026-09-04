//go:build test

package domain_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- Uniquely-prefixed helpers (shared package domain_test) ---

func newTestHandAndFoot() *domain.HandAndFoot {
	players := []*domain.HandAndFootPlayer{
		domain.NewHandAndFootPlayer(true),
		domain.NewHandAndFootPlayer(false),
		domain.NewHandAndFootPlayer(false),
		domain.NewHandAndFootPlayer(false),
	}
	return domain.NewHandAndFoot(domain.NewTrumpCardsWithDecks(4, 8), players, domain.DefaultHandAndFootConfig())
}

func newTestHandAndFootWithDifficulty(d domain.HandAndFootCpuDifficulty) *domain.HandAndFoot {
	players := []*domain.HandAndFootPlayer{
		domain.NewHandAndFootPlayer(true),
		domain.NewHandAndFootPlayer(false),
		domain.NewHandAndFootPlayer(false),
		domain.NewHandAndFootPlayer(false),
	}
	cfg := domain.DefaultHandAndFootConfig()
	cfg.CpuDifficulty = d
	return domain.NewHandAndFoot(domain.NewTrumpCardsWithDecks(4, 8), players, cfg)
}

func hafCard(d, v int) *domain.Card {
	return domain.NewCard(d, v, false)
}

func hafSetHand(p *domain.HandAndFootPlayer, cards ...*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func hafSetupDraw(g *domain.HandAndFoot, idx int) {
	g.SetPhase(domain.HandAndFootPhaseDraw)
	g.SetCurrentPlayerIdx(idx)
}

func hafSetupMeld(g *domain.HandAndFoot, idx int) {
	g.SetPhase(domain.HandAndFootPhaseMeld)
	g.SetCurrentPlayerIdx(idx)
}

func hafSetupDiscard(g *domain.HandAndFoot, idx int) {
	g.SetPhase(domain.HandAndFootPhaseDiscard)
	g.SetCurrentPlayerIdx(idx)
}

func hafCanastaMeld(rank int, count int, natural bool) *domain.CanastaMeld {
	cards := make([]*domain.Card, count)
	for i := range cards {
		cards[i] = hafCard(domain.CardDesignSpade, rank)
	}
	return &domain.CanastaMeld{Cards: cards, IsNatural: natural}
}

// --- Constructor & Reset ---

func TestNewHandAndFoot(t *testing.T) {
	g := newTestHandAndFoot()
	assert.Equal(t, -1, g.GetWinnerTeam())
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, 0, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, 4, g.GetPlayerCnt())
}

func TestHandAndFoot_TeamOf(t *testing.T) {
	assert.Equal(t, 0, domain.HandAndFootTeamOf(0))
	assert.Equal(t, 1, domain.HandAndFootTeamOf(1))
	assert.Equal(t, 0, domain.HandAndFootTeamOf(2))
	assert.Equal(t, 1, domain.HandAndFootTeamOf(3))
}

func TestHandAndFoot_Reset(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()

	assert.Equal(t, domain.HandAndFootPhaseDraw, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerTeam())

	// Each player: 22 hand cards (red 3s auto-laid + replaced), 13 foot cards.
	for i := 0; i < 4; i++ {
		p := g.GetPlayer(i)
		assert.Equal(t, domain.HandAndFootHandSize, p.GetCardsSize(), "player %d hand should have 22 cards", i)
		assert.Equal(t, domain.HandAndFootFootSize, p.GetFootSize(), "player %d foot should have 13 cards", i)
		for j := 0; j < p.GetCardsSize(); j++ {
			assert.False(t, domain.CanastaIsRed3(p.GetCard(j)), "player %d hand should not contain red 3s", i)
		}
		assert.False(t, p.GetInFoot())
	}

	assert.GreaterOrEqual(t, g.GetDiscardPileCount(), 1)

	// Total cards must be 216 across all locations.
	total := g.GetDrawPileCount() + g.GetDiscardPileCount()
	for i := 0; i < 4; i++ {
		total += g.GetPlayer(i).GetCardsSize() + g.GetPlayer(i).GetFootSize()
	}
	for t2 := 0; t2 < 2; t2++ {
		total += len(g.GetTeamRed3s(t2))
	}
	assert.Equal(t, 216, total, "total cards should be 216")
}

func TestHandAndFoot_Reset_DrawPileSize(t *testing.T) {
	// With no red 3s scenario the draw pile after deal is 216-140-1 = 75 (one to discard).
	// Red 3 replacements + multiple discard flips reduce it further, so just verify
	// the deal subtracted 140 hand/foot cards and at least one discard.
	g := newTestHandAndFoot()
	g.Reset()
	// 216 total - 140 (hands+feet) - red3s laid - discard flips = draw pile.
	dealt := 0
	for i := 0; i < 4; i++ {
		dealt += g.GetPlayer(i).GetCardsSize() + g.GetPlayer(i).GetFootSize()
	}
	assert.Equal(t, 140, dealt, "must deal 22+13 per player = 140")
	assert.LessOrEqual(t, g.GetDrawPileCount(), 75)
}

func TestHandAndFoot_Reset_ClearsAllState(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	g.SetGameEndFlag(true)

	g.Reset()
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, -1, g.GetWinnerTeam())
}

// --- Config ---

func TestHandAndFootConfig_Default(t *testing.T) {
	cfg := domain.DefaultHandAndFootConfig()
	assert.Equal(t, domain.HandAndFootCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 5000, cfg.PointLimit)
	assert.Equal(t, 1, cfg.RedCanastasToGoOut)
	assert.Equal(t, 1, cfg.BlackCanastasToGoOut)
}

func TestHandAndFootConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     domain.HandAndFootConfig
		wantErr bool
	}{
		{"valid default", domain.DefaultHandAndFootConfig(), false},
		{"valid easy", domain.HandAndFootConfig{CpuDifficulty: domain.HandAndFootCpuDifficultyEasy, PointLimit: 100, RedCanastasToGoOut: 0, BlackCanastasToGoOut: 0}, false},
		{"valid hard", domain.HandAndFootConfig{CpuDifficulty: domain.HandAndFootCpuDifficultyHard, PointLimit: 10000, RedCanastasToGoOut: 2, BlackCanastasToGoOut: 1}, false},
		{"invalid difficulty", domain.HandAndFootConfig{CpuDifficulty: 5, PointLimit: 5000, RedCanastasToGoOut: 1, BlackCanastasToGoOut: 1}, true},
		{"invalid point limit", domain.HandAndFootConfig{CpuDifficulty: domain.HandAndFootCpuDifficultyNormal, PointLimit: 0, RedCanastasToGoOut: 1, BlackCanastasToGoOut: 1}, true},
		{"invalid red canastas", domain.HandAndFootConfig{CpuDifficulty: domain.HandAndFootCpuDifficultyNormal, PointLimit: 5000, RedCanastasToGoOut: 99, BlackCanastasToGoOut: 1}, true},
		{"invalid black canastas", domain.HandAndFootConfig{CpuDifficulty: domain.HandAndFootCpuDifficultyNormal, PointLimit: 5000, RedCanastasToGoOut: 1, BlackCanastasToGoOut: -1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// --- Draw from stock ---

func TestHandAndFoot_PlayerDrawFromStock_PhaseGuard(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupMeld(g, 0)
	err := g.PlayerDrawFromStock()
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestHandAndFoot_PlayerDrawFromStock_HumanGuard(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDraw(g, 1)
	err := g.PlayerDrawFromStock()
	assert.True(t, errors.Is(err, domain.ErrNotHumanTurn))
}

func TestHandAndFoot_PlayerDrawFromStock_GameEndGuard(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	g.SetGameEndFlag(true)
	err := g.PlayerDrawFromStock()
	assert.True(t, errors.Is(err, domain.ErrGameEnded))
}

func TestHandAndFoot_PlayerDrawFromStock_Draws2(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDraw(g, 0)
	g.SetDrawPile([]*domain.Card{
		hafCard(domain.CardDesignSpade, 5),
		hafCard(domain.CardDesignHeart, 8),
		hafCard(domain.CardDesignClover, 9),
	})
	hafSetHand(g.GetPlayer(0), hafCard(domain.CardDesignSpade, 10))

	err := g.PlayerDrawFromStock()
	require.NoError(t, err)
	assert.Equal(t, domain.HandAndFootPhaseMeld, g.GetPhase())
	assert.Equal(t, 3, g.GetPlayer(0).GetCardsSize()) // 1 + 2 drawn
	assert.Equal(t, 1, g.GetDrawPileCount())
}

func TestHandAndFoot_PlayerDrawFromStock_EmptyDraw(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDraw(g, 0)
	g.SetDrawPile(nil)
	err := g.PlayerDrawFromStock()
	require.NoError(t, err)
	assert.True(t, g.GetPhase() == domain.HandAndFootPhaseRoundEnd || g.GetPhase() == domain.HandAndFootPhaseGameEnd)
}

// --- Draw from discard ---

func TestHandAndFoot_PlayerDrawFromDiscard_PhaseGuard(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupMeld(g, 0)
	err := g.PlayerDrawFromDiscard([]int{0, 1})
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestHandAndFoot_PlayerDrawFromDiscard_EmptyPile(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDraw(g, 0)
	g.SetDiscardPile(nil)
	err := g.PlayerDrawFromDiscard([]int{0, 1})
	assert.Error(t, err)
}

func TestHandAndFoot_PlayerDrawFromDiscard_Black3OnTop(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDraw(g, 0)
	g.SetDiscardPile([]*domain.Card{hafCard(domain.CardDesignSpade, 3)})
	err := g.PlayerDrawFromDiscard([]int{0, 1})
	assert.Error(t, err)
}

func TestHandAndFoot_PlayerDrawFromDiscard_WildOnTop(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDraw(g, 0)
	g.SetDiscardPile([]*domain.Card{hafCard(domain.CardDesignSpade, 2)})
	err := g.PlayerDrawFromDiscard([]int{0, 1})
	assert.Error(t, err)
}

func TestHandAndFoot_PlayerDrawFromDiscard_InvalidPairCount(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDraw(g, 0)
	g.SetDiscardPile([]*domain.Card{hafCard(domain.CardDesignSpade, 7)})
	err := g.PlayerDrawFromDiscard([]int{0})
	assert.Error(t, err)
}

func TestHandAndFoot_PlayerDrawFromDiscard_Success_Takes7(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDraw(g, 0)

	top := hafCard(domain.CardDesignSpade, 7)
	pile := []*domain.Card{}
	for i := 0; i < 9; i++ {
		pile = append(pile, hafCard(domain.CardDesignHeart, 5))
	}
	pile = append(pile, top) // top
	g.SetDiscardPile(pile)

	hafSetHand(g.GetPlayer(0),
		hafCard(domain.CardDesignHeart, 7),
		hafCard(domain.CardDesignDiamond, 7),
		hafCard(domain.CardDesignClover, 10))

	err := g.PlayerDrawFromDiscard([]int{0, 1})
	require.NoError(t, err)
	assert.Equal(t, domain.HandAndFootPhaseMeld, g.GetPhase())
	assert.True(t, g.GetDrewFromDiscard())
	// took at most 7 from the discard pile; 10 - 7 = 3 remain
	assert.Equal(t, 3, g.GetDiscardPileCount())
	// hand: 3 + 7 = 10
	assert.Equal(t, 10, g.GetPlayer(0).GetCardsSize())
}

func TestHandAndFoot_PlayerDrawFromDiscard_PairNotMatching(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDraw(g, 0)
	g.SetDiscardPile([]*domain.Card{hafCard(domain.CardDesignSpade, 7)})
	hafSetHand(g.GetPlayer(0), hafCard(domain.CardDesignHeart, 5), hafCard(domain.CardDesignDiamond, 5))
	err := g.PlayerDrawFromDiscard([]int{0, 1})
	assert.Error(t, err)
}

// --- Meld ---

func TestHandAndFoot_PlayerMeld_PhaseGuard(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDraw(g, 0)
	err := g.PlayerMeld([][]int{{0, 1, 2}})
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestHandAndFoot_PlayerMeld_EmptySkip(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupMeld(g, 0)
	err := g.PlayerMeld(nil)
	require.NoError(t, err)
	assert.Equal(t, domain.HandAndFootPhaseDiscard, g.GetPhase())
}

func TestHandAndFoot_PlayerMeld_ValidNewTeamMeld(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupMeld(g, 0)
	hafSetHand(g.GetPlayer(0),
		hafCard(domain.CardDesignSpade, 13),
		hafCard(domain.CardDesignHeart, 13),
		hafCard(domain.CardDesignDiamond, 13),
		hafCard(domain.CardDesignClover, 5))

	err := g.PlayerMeld([][]int{{0, 1, 2}})
	require.NoError(t, err)
	assert.Equal(t, 1, len(g.GetTeamMelds(0)))
	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, domain.HandAndFootPhaseDiscard, g.GetPhase())
}

func TestHandAndFoot_PlayerMeld_TooFewCards(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupMeld(g, 0)
	hafSetHand(g.GetPlayer(0), hafCard(domain.CardDesignSpade, 7), hafCard(domain.CardDesignHeart, 7))
	err := g.PlayerMeld([][]int{{0, 1}})
	assert.Error(t, err)
}

func TestHandAndFoot_PlayerMeld_TooManyWilds(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupMeld(g, 0)
	hafSetHand(g.GetPlayer(0),
		hafCard(domain.CardDesignSpade, 7),
		hafCard(domain.CardDesignJoker, 1),
		hafCard(domain.CardDesignSpade, 2),
		hafCard(domain.CardDesignHeart, 2),
		hafCard(domain.CardDesignDiamond, 2))
	err := g.PlayerMeld([][]int{{0, 1, 2, 3, 4}})
	assert.Error(t, err)
}

func TestHandAndFoot_PlayerMeld_Black3Rejected(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupMeld(g, 0)
	hafSetHand(g.GetPlayer(0),
		hafCard(domain.CardDesignSpade, 3),
		hafCard(domain.CardDesignClover, 3),
		hafCard(domain.CardDesignHeart, 5))
	err := g.PlayerMeld([][]int{{0, 1, 2}})
	assert.Error(t, err)
}

func TestHandAndFoot_PlayerMeld_AddToExistingTeamMeld(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupMeld(g, 0)
	// pre-existing team-0 meld of 9s (3 cards)
	g.SetTeamMelds(0, []*domain.CanastaMeld{hafCanastaMeld(9, 3, true)})
	hafSetHand(g.GetPlayer(0),
		hafCard(domain.CardDesignHeart, 9),
		hafCard(domain.CardDesignDiamond, 9),
		hafCard(domain.CardDesignClover, 5))
	err := g.PlayerMeld([][]int{{0, 1}})
	require.NoError(t, err)
	assert.Equal(t, 1, len(g.GetTeamMelds(0)))
	assert.Equal(t, 5, len(g.GetTeamMelds(0)[0].Cards))
}

func TestHandAndFoot_PlayerMeld_OutOfRangeIndex(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupMeld(g, 0)
	hafSetHand(g.GetPlayer(0), hafCard(domain.CardDesignSpade, 7))
	err := g.PlayerMeld([][]int{{0, 99}})
	assert.Error(t, err)
}

// --- Hand -> Foot transition ---

func TestHandAndFoot_HandToFootTransition_OnMeldEmptyHand(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupMeld(g, 0)
	p := g.GetPlayer(0)
	p.SetFoot([]*domain.Card{hafCard(domain.CardDesignSpade, 10), hafCard(domain.CardDesignHeart, 11)})
	hafSetHand(p,
		hafCard(domain.CardDesignSpade, 13),
		hafCard(domain.CardDesignHeart, 13),
		hafCard(domain.CardDesignDiamond, 13))

	err := g.PlayerMeld([][]int{{0, 1, 2}})
	require.NoError(t, err)
	// hand emptied → foot picked up
	assert.True(t, p.GetInFoot())
	assert.Equal(t, 0, p.GetFootSize())
	assert.Equal(t, 2, p.GetCardsSize())
	// still in meld phase (turn continues with foot)
	assert.Equal(t, domain.HandAndFootPhaseMeld, g.GetPhase())
}

// --- Canasta detection ---

func TestHandAndFoot_CanastaDetection_RedVsBlack(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	red := hafCanastaMeld(8, 7, true)
	black := &domain.CanastaMeld{
		Cards: []*domain.Card{
			hafCard(domain.CardDesignSpade, 9), hafCard(domain.CardDesignHeart, 9),
			hafCard(domain.CardDesignDiamond, 9), hafCard(domain.CardDesignClover, 9),
			hafCard(domain.CardDesignSpade, 9), hafCard(domain.CardDesignHeart, 9),
			hafCard(domain.CardDesignJoker, 1), // wild → dirty
		},
		IsNatural: false,
	}
	g.SetTeamMelds(0, []*domain.CanastaMeld{red, black})
	assert.True(t, red.IsCanasta())
	assert.True(t, red.IsNatural)
	assert.True(t, black.IsCanasta())
	assert.False(t, black.IsNatural)
}

// --- Go out gating ---

func TestHandAndFoot_PlayerGoOut_NotInFootFails(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDiscard(g, 0)
	p := g.GetPlayer(0)
	p.SetInFoot(false)
	g.SetTeamMelds(0, []*domain.CanastaMeld{hafCanastaMeld(7, 7, true), hafCanastaMeld(8, 7, false)})
	hafSetHand(p, hafCard(domain.CardDesignSpade, 5))
	err := g.PlayerGoOut()
	assert.Error(t, err)
}

func TestHandAndFoot_PlayerGoOut_MissingCanastasFails(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDiscard(g, 0)
	p := g.GetPlayer(0)
	p.SetInFoot(true)
	// only a red canasta, no black canasta
	g.SetTeamMelds(0, []*domain.CanastaMeld{hafCanastaMeld(7, 7, true)})
	hafSetHand(p, hafCard(domain.CardDesignSpade, 5))
	err := g.PlayerGoOut()
	assert.Error(t, err)
}

func TestHandAndFoot_PlayerGoOut_Success(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDiscard(g, 0)
	p := g.GetPlayer(0)
	p.SetInFoot(true)
	g.SetTeamMelds(0, []*domain.CanastaMeld{hafCanastaMeld(7, 7, true), hafCanastaMeld(8, 7, false)})
	hafSetHand(p, hafCard(domain.CardDesignClover, 5))
	err := g.PlayerGoOut()
	require.NoError(t, err)
	assert.True(t, g.GetPhase() == domain.HandAndFootPhaseRoundEnd || g.GetPhase() == domain.HandAndFootPhaseGameEnd)
}

func TestHandAndFoot_PlayerGoOut_PhaseGuard(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDraw(g, 0)
	err := g.PlayerGoOut()
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

// --- Discard ---

func TestHandAndFoot_PlayerDiscard_PhaseGuard(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDraw(g, 0)
	err := g.PlayerDiscard(0)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestHandAndFoot_PlayerDiscard_InvalidIndex(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDiscard(g, 0)
	err := g.PlayerDiscard(999)
	assert.Error(t, err)
}

func TestHandAndFoot_PlayerDiscard_Red3Rejected(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDiscard(g, 0)
	hafSetHand(g.GetPlayer(0), hafCard(domain.CardDesignHeart, 3))
	err := g.PlayerDiscard(0)
	assert.Error(t, err)
}

func TestHandAndFoot_PlayerDiscard_Success_AdvancesTurn(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDiscard(g, 0)
	g.SetDrawPile([]*domain.Card{hafCard(domain.CardDesignSpade, 9)})
	hafSetHand(g.GetPlayer(0), hafCard(domain.CardDesignSpade, 7), hafCard(domain.CardDesignHeart, 8))
	before := g.GetDiscardPileCount()
	err := g.PlayerDiscard(0)
	require.NoError(t, err)
	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, before+1, g.GetDiscardPileCount())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	assert.Equal(t, domain.HandAndFootPhaseDraw, g.GetPhase())
}

func TestHandAndFoot_PlayerDiscard_WildFreezes(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDiscard(g, 0)
	g.SetIsFrozen(false)
	g.SetDrawPile([]*domain.Card{hafCard(domain.CardDesignSpade, 9)})
	hafSetHand(g.GetPlayer(0), hafCard(domain.CardDesignSpade, 2), hafCard(domain.CardDesignHeart, 8))
	err := g.PlayerDiscard(0)
	require.NoError(t, err)
	assert.True(t, g.GetIsFrozen())
}

func TestHandAndFoot_PlayerDiscard_EmptiesHandPicksUpFoot(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDiscard(g, 0)
	g.SetDrawPile([]*domain.Card{hafCard(domain.CardDesignSpade, 9)})
	p := g.GetPlayer(0)
	p.SetFoot([]*domain.Card{hafCard(domain.CardDesignClover, 10)})
	hafSetHand(p, hafCard(domain.CardDesignSpade, 7))
	err := g.PlayerDiscard(0)
	require.NoError(t, err)
	assert.True(t, p.GetInFoot())
}

// --- SkipMeld ---

func TestHandAndFoot_PlayerSkipMeld(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupMeld(g, 0)
	err := g.PlayerSkipMeld()
	require.NoError(t, err)
	assert.Equal(t, domain.HandAndFootPhaseDiscard, g.GetPhase())
}

func TestHandAndFoot_PlayerSkipMeld_PhaseGuard(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDraw(g, 0)
	err := g.PlayerSkipMeld()
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

// --- NextRound ---

func TestHandAndFoot_NextRound_WrongPhase(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	g.NextRound()
	assert.Equal(t, domain.HandAndFootPhaseDraw, g.GetPhase())
}

func TestHandAndFoot_NextRound_FromRoundEnd(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	g.SetPhase(domain.HandAndFootPhaseRoundEnd)
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, domain.HandAndFootPhaseDraw, g.GetPhase())
	for i := 0; i < 4; i++ {
		assert.Equal(t, domain.HandAndFootHandSize, g.GetPlayer(i).GetCardsSize())
		assert.Equal(t, domain.HandAndFootFootSize, g.GetPlayer(i).GetFootSize())
	}
}

// --- Scoring ---

func TestHandAndFoot_Scoring_RoundEnd(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDiscard(g, 0)
	// Reset() deals 140 cards and auto-lays any red 3s into the team piles, and
	// leaves each player holding a 13-card foot. Clear ALL of that dealt state so
	// the scoring scenario below is deterministic (otherwise red-3 bonuses and
	// uncleared-foot penalties leak in and make this test shuffle-dependent).
	g.SetTeamRed3s(0, nil)
	g.SetTeamRed3s(1, nil)
	for i := 0; i < 4; i++ {
		hafSetHand(g.GetPlayer(i))
		g.GetPlayer(i).SetFoot(nil)
		g.GetPlayer(i).SetInFoot(true)
	}
	p := g.GetPlayer(0)
	// team 0: red canasta of 8s (7×10=70 + 500), black canasta of 9s (6 nat ×10 + 1 joker 50 = 110 + 300)
	red := hafCanastaMeld(8, 7, true)
	black := &domain.CanastaMeld{
		Cards: []*domain.Card{
			hafCard(domain.CardDesignSpade, 9), hafCard(domain.CardDesignHeart, 9),
			hafCard(domain.CardDesignDiamond, 9), hafCard(domain.CardDesignClover, 9),
			hafCard(domain.CardDesignSpade, 9), hafCard(domain.CardDesignHeart, 9),
			hafCard(domain.CardDesignJoker, 1),
		},
		IsNatural: false,
	}
	g.SetTeamMelds(0, []*domain.CanastaMeld{red, black})
	hafSetHand(p, hafCard(domain.CardDesignClover, 5)) // discard this 1 card, go out

	err := g.PlayerGoOut()
	require.NoError(t, err)
	// Team0 cumulative: 8s=70+500=570 ; 9s=60+50=110 +300=410 ; +goout 100 = 1080
	assert.Equal(t, 1080, g.GetPlayer(0).GetCumulativeScore())
	assert.Equal(t, 1080, g.GetPlayer(2).GetCumulativeScore()) // same team
	assert.Equal(t, 0, g.GetPlayer(1).GetCumulativeScore())    // team1, nothing
}

func TestHandAndFoot_Scoring_Red3Bonus(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDiscard(g, 0)
	p := g.GetPlayer(0)
	p.SetInFoot(true)
	g.SetTeamMelds(0, []*domain.CanastaMeld{hafCanastaMeld(7, 7, true), hafCanastaMeld(10, 7, false)})
	// 2 red 3s for team 0 (via setting? use round end path)
	for i := 1; i < 4; i++ {
		hafSetHand(g.GetPlayer(i))
		g.GetPlayer(i).SetFoot(nil)
	}
	hafSetHand(p)
	p.SetFoot(nil)
	// emptied hand & in foot; trigger via CpuPlay? Use direct go out (0 cards).
	err := g.PlayerGoOut()
	require.NoError(t, err)
	assert.Greater(t, g.GetPlayer(0).GetCumulativeScore(), 0)
}

// --- CpuPlay ---

func TestHandAndFoot_CpuPlay_Draw(t *testing.T) {
	for _, diff := range []domain.HandAndFootCpuDifficulty{
		domain.HandAndFootCpuDifficultyEasy,
		domain.HandAndFootCpuDifficultyNormal,
		domain.HandAndFootCpuDifficultyHard,
	} {
		t.Run(fmt.Sprintf("difficulty_%d", diff), func(t *testing.T) {
			g := newTestHandAndFootWithDifficulty(diff)
			g.Reset()
			hafSetupDraw(g, 1)
			g.CpuPlay()
			assert.NotEqual(t, domain.HandAndFootPhaseDraw, g.GetPhase())
		})
	}
}

// **配りに依存させない (#6641)。** hafSetupMeld はフェーズと手番を立てるだけで
// 手札には触らないので、CPU がメルドできるかどうかは `Reset()` が配った手札
// 次第だった。メルドで手札を出し切った局では `cpuMeld` が `enterFootIfEmpty`
// で早期 return し、**phase は Meld のまま**残る ── `NotEqual(Meld)` はそこで
// 落ちる。CI で無関係な PR を赤くしていたのはこれ。手札を組んでから測る。
func TestHandAndFoot_CpuPlay_Meld(t *testing.T) {
	t.Run("melds the set it holds and moves on to the discard", func(t *testing.T) {
		g := newTestHandAndFoot()
		g.Reset()
		hafSetupMeld(g, 1)
		// 7 が3枚でメルド1組。残る2枚は同ランクにならず、ワイルド (2/Joker)
		// でもないので、手札は空にならない = フット取り込みは起きない。
		hafSetHand(g.GetPlayer(1),
			hafCard(domain.CardDesignSpade, 7),
			hafCard(domain.CardDesignHeart, 7),
			hafCard(domain.CardDesignClover, 7),
			hafCard(domain.CardDesignDiamond, 5),
			hafCard(domain.CardDesignSpade, 9),
		)

		g.CpuPlay()

		assert.Equal(t, domain.HandAndFootPhaseDiscard, g.GetPhase())
		assert.Equal(t, 2, g.GetPlayer(1).GetCardsSize(), "メルドした3枚が手札から抜けていない")
	})

	// 負のコントロール: 出せる組が1つも無くても手番は進む。ここで止まると
	// CPU ループが回り続ける。
	t.Run("still moves on when nothing can be melded", func(t *testing.T) {
		g := newTestHandAndFoot()
		g.Reset()
		hafSetupMeld(g, 1)
		// 同ランク3枚も、2枚+ワイルドも作れない組み合わせ。
		hafSetHand(g.GetPlayer(1),
			hafCard(domain.CardDesignSpade, 5),
			hafCard(domain.CardDesignHeart, 6),
			hafCard(domain.CardDesignClover, 8),
			hafCard(domain.CardDesignDiamond, 9),
			hafCard(domain.CardDesignSpade, 10),
		)

		g.CpuPlay()

		assert.Equal(t, domain.HandAndFootPhaseDiscard, g.GetPhase())
		assert.Equal(t, 5, g.GetPlayer(1).GetCardsSize(), "出せないはずの札が出ている")
	})
}

func TestHandAndFoot_CpuPlay_Discard(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDiscard(g, 1)
	if g.GetPlayer(1).GetCardsSize() == 0 {
		g.GetPlayer(1).AddCard(hafCard(domain.CardDesignSpade, 7))
	}
	g.CpuPlay()
	// turn should advance (or round end)
	assert.NotEqual(t, 1, g.GetCurrentPlayerIdx())
}

func TestHandAndFoot_CpuPlay_HumanTurn_Noop(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	hafSetupDraw(g, 0)
	g.CpuPlay()
	assert.Equal(t, domain.HandAndFootPhaseDraw, g.GetPhase())
}

func TestHandAndFoot_CpuPlay_GameEnded_Noop(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	g.SetGameEndFlag(true)
	g.CpuPlay()
}

// --- Full CPU game terminates ---

func TestHandAndFoot_FullCpuGame_Terminates(t *testing.T) {
	players := []*domain.HandAndFootPlayer{
		domain.NewHandAndFootPlayer(false),
		domain.NewHandAndFootPlayer(false),
		domain.NewHandAndFootPlayer(false),
		domain.NewHandAndFootPlayer(false),
	}
	cfg := domain.DefaultHandAndFootConfig()
	cfg.PointLimit = 500 // low limit so the game ends quickly
	g := domain.NewHandAndFoot(domain.NewTrumpCardsWithDecks(4, 8), players, cfg)
	g.Reset()

	const maxIter = 200000
	iter := 0
	for !g.GetGameEndFlag() && iter < maxIter {
		iter++
		phase := g.GetPhase()
		if phase == domain.HandAndFootPhaseRoundEnd {
			g.NextRound()
			continue
		}
		if phase == domain.HandAndFootPhaseGameEnd {
			break
		}
		g.CpuPlay()
	}
	assert.Less(t, iter, maxIter, "full CPU game should terminate within %d iterations", maxIter)
	assert.True(t, g.GetGameEndFlag())
	winner := g.GetWinnerTeam()
	require.GreaterOrEqual(t, winner, 0)
	// winning team cumulative >= PointLimit
	rep := g.GetPlayer(g.GetWinnerIdx())
	assert.GreaterOrEqual(t, rep.GetCumulativeScore(), cfg.PointLimit)
}

// --- Getters & State ---

func TestHandAndFoot_GetPlayer_OutOfBounds(t *testing.T) {
	g := newTestHandAndFoot()
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(10))
}

func TestHandAndFoot_GetTeamMelds_OutOfBounds(t *testing.T) {
	g := newTestHandAndFoot()
	assert.Nil(t, g.GetTeamMelds(-1))
	assert.Nil(t, g.GetTeamMelds(5))
	assert.Nil(t, g.GetTeamRed3s(-1))
}

func TestHandAndFoot_IsHumanTurn(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(1)
	assert.False(t, g.IsHumanTurn())
}

func TestHandAndFoot_GetSetConfig(t *testing.T) {
	g := newTestHandAndFoot()
	cfg := domain.HandAndFootConfig{CpuDifficulty: domain.HandAndFootCpuDifficultyHard, PointLimit: 10000, RedCanastasToGoOut: 2, BlackCanastasToGoOut: 2}
	g.SetConfig(cfg)
	assert.Equal(t, cfg, g.GetConfig())
}

// --- CanastaMeld reuse ---

func TestHandAndFootMeld_IsCanasta(t *testing.T) {
	assert.False(t, hafCanastaMeld(7, 6, true).IsCanasta())
	assert.True(t, hafCanastaMeld(7, 7, true).IsCanasta())
}

// --- HandAndFootPlayer ---

func TestHandAndFootPlayer_ResetRound(t *testing.T) {
	p := domain.NewHandAndFootPlayer(true)
	p.AddCard(hafCard(domain.CardDesignSpade, 5))
	p.AddFootCard(hafCard(domain.CardDesignHeart, 9))
	p.SetInFoot(true)
	p.SetRoundScore(100)

	p.ResetRound()
	assert.Equal(t, 0, p.GetCardsSize())
	assert.Equal(t, 0, p.GetFootSize())
	assert.False(t, p.GetInFoot())
	assert.Equal(t, 0, p.GetRoundScore())
}

// --- JSON Serialization ---

func TestHandAndFoot_JSON_RoundTrip(t *testing.T) {
	g := newTestHandAndFoot()
	g.Reset()
	g.SetTeamMelds(0, []*domain.CanastaMeld{hafCanastaMeld(7, 3, true)})

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var g2 domain.HandAndFoot
	err = json.Unmarshal(data, &g2)
	require.NoError(t, err)

	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetRoundNumber(), g2.GetRoundNumber())
	assert.Equal(t, g.GetCurrentPlayerIdx(), g2.GetCurrentPlayerIdx())
	assert.Equal(t, g.GetIsFrozen(), g2.GetIsFrozen())
	assert.Equal(t, g.GetGameEndFlag(), g2.GetGameEndFlag())
	assert.Equal(t, g.GetWinnerTeam(), g2.GetWinnerTeam())
	assert.Equal(t, g.GetDrawPileCount(), g2.GetDrawPileCount())
	assert.Equal(t, g.GetDiscardPileCount(), g2.GetDiscardPileCount())
	assert.Equal(t, len(g.GetTeamMelds(0)), len(g2.GetTeamMelds(0)))
}

func TestHandAndFootPlayer_JSON_RoundTrip(t *testing.T) {
	p := domain.NewHandAndFootPlayer(true)
	p.AddCard(hafCard(domain.CardDesignSpade, 7))
	p.AddFootCard(hafCard(domain.CardDesignHeart, 9))
	p.SetInFoot(true)
	p.SetRoundScore(50)

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var p2 domain.HandAndFootPlayer
	err = json.Unmarshal(data, &p2)
	require.NoError(t, err)

	assert.Equal(t, p.GetCardsSize(), p2.GetCardsSize())
	assert.Equal(t, p.GetFootSize(), p2.GetFootSize())
	assert.Equal(t, p.GetInFoot(), p2.GetInFoot())
	assert.Equal(t, p.GetRoundScore(), p2.GetRoundScore())
}

func TestHandAndFoot_UnmarshalJSON_RejectsInvalid(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"invalid phase", `{"ps":99,"wt":-1,"cf":{"cd":1,"pl":5000,"rc":1,"bc":1}}`},
		{"invalid winner team", `{"ps":0,"wt":5,"cf":{"cd":1,"pl":5000,"rc":1,"bc":1}}`},
		{"invalid config difficulty", `{"ps":0,"wt":-1,"cf":{"cd":9,"pl":5000,"rc":1,"bc":1}}`},
		{"bad json", `{not json`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var g domain.HandAndFoot
			err := json.Unmarshal([]byte(tt.data), &g)
			assert.Error(t, err)
		})
	}
}

func TestHandAndFoot_UnmarshalJSON_InvalidPlayerCount(t *testing.T) {
	// One player but expecting 4.
	data := `{"ps":0,"wt":-1,"ci":0,"cf":{"cd":1,"pl":5000,"rc":1,"bc":1},"pl":[{"gp":{"ih":true,"ca":[]}}]}`
	var g domain.HandAndFoot
	err := json.Unmarshal([]byte(data), &g)
	assert.Error(t, err)
}

// **Web の canastaMinMeld と同じ帯であること (#4836)。**
func TestCanastaMinMeld(t *testing.T) {
	assert.Equal(t, 15, domain.CanastaMinMeld(-50))
	assert.Equal(t, 50, domain.CanastaMinMeld(0))
	assert.Equal(t, 50, domain.CanastaMinMeld(1499))
	assert.Equal(t, 90, domain.CanastaMinMeld(1500))
	assert.Equal(t, 90, domain.CanastaMinMeld(2999))
	assert.Equal(t, 120, domain.CanastaMinMeld(3000))
}

// **「上がれる」表示と実際に上がれるかがずれないこと。**GetGoOutStatus は
// canGoOut と同じ判定を使う (#4836)。
func TestHandAndFoot_GetGoOutStatus(t *testing.T) {
	g := domain.NewDefaultHandAndFoot()
	g.Reset()

	st := g.GetGoOutStatus(0)
	assert.False(t, st.InFoot, "配った直後はフットに入っていない")
	assert.False(t, st.CanGoOut())
	assert.Positive(t, st.RedRequired)
	assert.Positive(t, st.BlackReq)

	// 範囲外は空 (= 上がれない)。
	assert.False(t, g.GetGoOutStatus(-1).CanGoOut())
	assert.False(t, g.GetGoOutStatus(99).CanGoOut())

	// 3 条件が揃えば CanGoOut は真。
	full := domain.HandAndFootGoOutStatus{InFoot: true, RedCanastas: 1, RedRequired: 1, BlackCanasta: 1, BlackReq: 1}
	assert.True(t, full.CanGoOut())
	noFoot := full
	noFoot.InFoot = false
	assert.False(t, noFoot.CanGoOut())
}
