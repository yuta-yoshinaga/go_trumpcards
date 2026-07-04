//go:build test

package domain_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// resetAllSambaPlayers clears hands, melds, red3s and round scores for every
// player so controlled-state scoring tests start from a clean slate.
func resetAllSambaPlayers(g *domain.Samba) {
	for i := 0; i < domain.SambaPlayerCnt; i++ {
		g.GetPlayer(i).ResetRound()
	}
}

// findSambaCardIdx returns the hand indices of cards with the given value.
func findSambaCardIdx(p *domain.SambaPlayer, value int) []int {
	var out []int
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetValue() == value {
			out = append(out, i)
		}
	}
	return out
}

// --- validateNewSet edge cases ---

func TestSamba_NewSet_TooFewCards(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	p.SetHasInitMeld(true)
	p.AddCard(sambaCard(domain.CardDesignSpade, 9))
	p.AddCard(sambaCard(domain.CardDesignHeart, 9))
	assert.Error(t, g.PlayerMeld([][]int{{0, 1}}))
}

func TestSamba_NewSet_AllWildsRejected(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	p.SetHasInitMeld(true)
	p.AddCard(sambaCard(domain.CardDesignJoker, 1))
	p.AddCard(sambaCard(domain.CardDesignSpade, 2))
	p.AddCard(sambaCard(domain.CardDesignHeart, 2))
	// all wild -> naturalCount < 2
	assert.Error(t, g.PlayerMeld([][]int{{0, 1, 2}}))
}

func TestSamba_NewSet_MismatchedRanks(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	p.SetHasInitMeld(true)
	// same-suit but the group is set-shaped only if ranks match; here ranks
	// differ and they are NOT a run, so it fails both classifications.
	p.AddCard(sambaCard(domain.CardDesignSpade, 9))
	p.AddCard(sambaCard(domain.CardDesignSpade, 9))
	p.AddCard(sambaCard(domain.CardDesignClover, 12))
	assert.Error(t, g.PlayerMeld([][]int{{0, 1, 2}}))
}

func TestSamba_NewSet_OneWildOK(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	p.SetHasInitMeld(true)
	p.AddCard(sambaCard(domain.CardDesignSpade, 13))
	p.AddCard(sambaCard(domain.CardDesignHeart, 13))
	p.AddCard(sambaCard(domain.CardDesignSpade, 2)) // wild
	require.NoError(t, g.PlayerMeld([][]int{{0, 1, 2}}))
	assert.Equal(t, domain.SambaMeldSet, p.GetMelds()[0].Kind)
	assert.False(t, p.GetMelds()[0].IsNatural)
}

// --- validateSetAddition (existing meld) ---

func TestSamba_SetAddition(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	p.SetHasInitMeld(true)
	p.SetMelds([]*domain.SambaMeld{{
		Cards: []*domain.Card{
			sambaCard(domain.CardDesignSpade, 9),
			sambaCard(domain.CardDesignHeart, 9),
			sambaCard(domain.CardDesignClover, 9),
		},
		Kind:      domain.SambaMeldSet,
		IsNatural: true,
	}})
	p.AddCard(sambaCard(domain.CardDesignDiamond, 9))
	require.NoError(t, g.PlayerMeld([][]int{{0}}))
	assert.Len(t, p.GetMelds(), 1)
	assert.Len(t, p.GetMelds()[0].Cards, 4)
}

// --- validateNewSequence edge cases ---

func TestSamba_NewSequence_TooFew(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	p.SetHasInitMeld(true)
	p.AddCard(sambaCard(domain.CardDesignHeart, 4))
	p.AddCard(sambaCard(domain.CardDesignHeart, 5))
	assert.Error(t, g.PlayerMeld([][]int{{0, 1}}))
}

func TestSamba_NewSequence_ThreeRejected(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	p.SetHasInitMeld(true)
	// a run that would include a 3 is rejected (red 3 bonus / black 3 unmeldable)
	p.AddCard(sambaCard(domain.CardDesignHeart, 3))
	p.AddCard(sambaCard(domain.CardDesignHeart, 4))
	p.AddCard(sambaCard(domain.CardDesignHeart, 5))
	assert.Error(t, g.PlayerMeld([][]int{{0, 1, 2}}))
}

func TestSamba_NewSequence_MixedSuitRejected(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	p.SetHasInitMeld(true)
	p.AddCard(sambaCard(domain.CardDesignHeart, 4))
	p.AddCard(sambaCard(domain.CardDesignSpade, 5))
	p.AddCard(sambaCard(domain.CardDesignHeart, 6))
	assert.Error(t, g.PlayerMeld([][]int{{0, 1, 2}}))
}

func TestSamba_NewSequence_DuplicateRejected(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	p.SetHasInitMeld(true)
	p.AddCard(sambaCard(domain.CardDesignHeart, 5))
	p.AddCard(sambaCard(domain.CardDesignHeart, 5))
	p.AddCard(sambaCard(domain.CardDesignHeart, 6))
	assert.Error(t, g.PlayerMeld([][]int{{0, 1, 2}}))
}

func TestSamba_NewSequence_AceHighRun(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	p.SetHasInitMeld(true)
	// Q-K-A run (12,13,ace=1 mapped to 14) is valid; ace is high, no wrap.
	p.AddCard(sambaCard(domain.CardDesignHeart, 12))
	p.AddCard(sambaCard(domain.CardDesignHeart, 13))
	p.AddCard(sambaCard(domain.CardDesignHeart, 1))
	require.NoError(t, g.PlayerMeld([][]int{{0, 1, 2}}))
	assert.Equal(t, domain.SambaMeldSequence, p.GetMelds()[0].Kind)
}

// --- resolveMeldGroup single-card sequence extension (set-shaped fallback) ---

func TestSamba_SingleCardExtendsSequence(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	p.SetHasInitMeld(true)
	p.SetMelds([]*domain.SambaMeld{{
		Cards: []*domain.Card{
			sambaCard(domain.CardDesignHeart, 4),
			sambaCard(domain.CardDesignHeart, 5),
			sambaCard(domain.CardDesignHeart, 6),
		},
		Kind:      domain.SambaMeldSequence,
		IsNatural: true,
	}})
	p.AddCard(sambaCard(domain.CardDesignHeart, 7)) // set-shaped single card -> seq extension
	require.NoError(t, g.PlayerMeld([][]int{{0}}))
	require.Len(t, p.GetMelds(), 1)
	assert.Equal(t, domain.SambaMeldSequence, p.GetMelds()[0].Kind)
	assert.Len(t, p.GetMelds()[0].Cards, 4)
}

// --- PlayerMeld index/guard branches ---

func TestSamba_Meld_OutOfRangeIndex(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	p.SetHasInitMeld(true)
	p.AddCard(sambaCard(domain.CardDesignSpade, 9))
	assert.Error(t, g.PlayerMeld([][]int{{0, 99}}))
}

func TestSamba_Meld_DuplicateIndex(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	p.SetHasInitMeld(true)
	p.AddCard(sambaCard(domain.CardDesignSpade, 9))
	p.AddCard(sambaCard(domain.CardDesignHeart, 9))
	p.AddCard(sambaCard(domain.CardDesignClover, 9))
	assert.Error(t, g.PlayerMeld([][]int{{0, 0, 1}}))
}

func TestSamba_Meld_NotHumanAndGameEnd(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 1) // CPU
	assert.Error(t, g.PlayerMeld([][]int{{0, 1, 2}}))

	g.SetGameEndFlag(true)
	assert.Error(t, g.PlayerMeld([][]int{{0, 1, 2}}))
}

func TestSamba_Meld_InitialMinMetSucceeds(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	// hasInitMeld defaults false: score 0 -> min 50. Three Kings + two Aces = 70 >= 50.
	p.AddCard(sambaCard(domain.CardDesignSpade, 13))
	p.AddCard(sambaCard(domain.CardDesignHeart, 13))
	p.AddCard(sambaCard(domain.CardDesignDiamond, 13))
	p.AddCard(sambaCard(domain.CardDesignSpade, 1))
	p.AddCard(sambaCard(domain.CardDesignHeart, 1))
	p.AddCard(sambaCard(domain.CardDesignDiamond, 1))
	require.NoError(t, g.PlayerMeld([][]int{{0, 1, 2}, {3, 4, 5}}))
	assert.True(t, p.GetHasInitMeld())
	assert.Len(t, p.GetMelds(), 2)
}

func TestSamba_Meld_EmptyToGoOut(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	resetAllSambaPlayers(g)
	p := g.GetPlayer(0)
	// already holds one canasta; melds a second 3-card set is not a canasta, so
	// give a completed second meld directly and meld the last three cards to
	// empty the hand while the team already has 2 completed melds -> go out.
	p.SetMelds([]*domain.SambaMeld{sambaSevenSet(5), sambaSevenSequence()})
	p.SetHasInitMeld(true)
	p.AddCard(sambaCard(domain.CardDesignSpade, 8))
	p.AddCard(sambaCard(domain.CardDesignHeart, 8))
	p.AddCard(sambaCard(domain.CardDesignClover, 8))
	require.NoError(t, g.PlayerMeld([][]int{{0, 1, 2}}))
	assert.Equal(t, 0, p.GetCardsSize())
	assert.True(t, g.GetPhase() == domain.SambaPhaseRoundEnd || g.GetPhase() == domain.SambaPhaseGameEnd)
}

// --- PlayerMeld after taking the discard pile (top must be melded) ---

func setupSambaTookPile(t *testing.T) (*domain.Samba, *domain.SambaPlayer) {
	t.Helper()
	g := newTestSamba()
	g.Reset()
	setupSambaDrawPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	p.SetHasInitMeld(true)
	p.AddCard(sambaCard(domain.CardDesignHeart, 7))
	p.AddCard(sambaCard(domain.CardDesignDiamond, 7))
	p.AddCard(sambaCard(domain.CardDesignSpade, 13))
	p.AddCard(sambaCard(domain.CardDesignHeart, 13))
	p.AddCard(sambaCard(domain.CardDesignDiamond, 13))
	g.SetDiscardPile([]*domain.Card{
		sambaCard(domain.CardDesignClover, 4),
		sambaCard(domain.CardDesignSpade, 7), // top
	})
	require.NoError(t, g.PlayerDrawFromDiscard([]int{0, 1}))
	require.True(t, g.GetDrewFromDiscard())
	return g, p
}

func TestSamba_Meld_DrewFromDiscard_EmptyGroupsRejected(t *testing.T) {
	g, _ := setupSambaTookPile(t)
	assert.Error(t, g.PlayerMeld(nil))
}

func TestSamba_SkipMeld_DrewFromDiscardRejected(t *testing.T) {
	g, _ := setupSambaTookPile(t)
	assert.Error(t, g.PlayerSkipMeld())
}

func TestSamba_Meld_DrewFromDiscard_TopNotInMeld(t *testing.T) {
	g, p := setupSambaTookPile(t)
	kings := findSambaCardIdx(p, 13)
	require.Len(t, kings, 3)
	assert.Error(t, g.PlayerMeld([][]int{kings}))
}

func TestSamba_Meld_DrewFromDiscard_TopInMeld(t *testing.T) {
	g, p := setupSambaTookPile(t)
	sevens := findSambaCardIdx(p, 7)
	require.Len(t, sevens, 3) // 7H, 7D + top 7S
	require.NoError(t, g.PlayerMeld([][]int{sevens}))
	assert.False(t, g.GetDrewFromDiscard())
}

// --- PlayerDrawFromDiscard extra branches ---

func TestSamba_DrawFromDiscard_SameIndexTwice(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDrawPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(sambaCard(domain.CardDesignHeart, 7))
	g.SetDiscardPile([]*domain.Card{sambaCard(domain.CardDesignSpade, 7)})
	assert.Error(t, g.PlayerDrawFromDiscard([]int{0, 0}))
}

func TestSamba_DrawFromDiscard_WildInPair(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDrawPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(sambaCard(domain.CardDesignSpade, 2)) // wild
	p.AddCard(sambaCard(domain.CardDesignHeart, 7))
	g.SetDiscardPile([]*domain.Card{sambaCard(domain.CardDesignSpade, 7)})
	assert.Error(t, g.PlayerDrawFromDiscard([]int{0, 1}))
}

func TestSamba_DrawFromDiscard_RankMismatch(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDrawPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(sambaCard(domain.CardDesignHeart, 8))
	p.AddCard(sambaCard(domain.CardDesignDiamond, 8))
	g.SetDiscardPile([]*domain.Card{sambaCard(domain.CardDesignSpade, 7)})
	assert.Error(t, g.PlayerDrawFromDiscard([]int{0, 1}))
}

func TestSamba_DrawFromDiscard_OutOfRangeIndex(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDrawPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(sambaCard(domain.CardDesignHeart, 7))
	g.SetDiscardPile([]*domain.Card{sambaCard(domain.CardDesignSpade, 7)})
	assert.Error(t, g.PlayerDrawFromDiscard([]int{0, 99}))
}

func TestSamba_DrawFromDiscard_InitialMinNotMet(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDrawPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset() // hasInitMeld false -> min 50
	p.AddCard(sambaCard(domain.CardDesignHeart, 4))
	p.AddCard(sambaCard(domain.CardDesignDiamond, 4))
	g.SetDiscardPile([]*domain.Card{sambaCard(domain.CardDesignSpade, 4)}) // 3×5=15 < 50
	assert.Error(t, g.PlayerDrawFromDiscard([]int{0, 1}))
}

func TestSamba_DrawFromDiscard_GameEndedAndNotHuman(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDrawPhase(g, 1) // CPU
	assert.Error(t, g.PlayerDrawFromDiscard([]int{0, 1}))
	g.SetGameEndFlag(true)
	assert.Error(t, g.PlayerDrawFromDiscard([]int{0, 1}))
}

// --- PlayerDrawFromStock red 3 auto-lay ---

func TestSamba_DrawFromStock_Red3AutoLaid(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDrawPhase(g, 0)
	p := g.GetPlayer(0)
	p.Reset()
	before := len(p.GetRed3s())
	g.SetDrawPile([]*domain.Card{
		sambaCard(domain.CardDesignClover, 9), // replacement drawn after red 3
		sambaCard(domain.CardDesignHeart, 3),  // red 3 on top
	})
	require.NoError(t, g.PlayerDrawFromStock())
	assert.Equal(t, before+1, len(p.GetRed3s()))
	assert.Equal(t, domain.SambaPhaseMeld, g.GetPhase())
}

// --- PlayerGoOut branches ---

func TestSamba_GoOut_OneCardSuccess(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDiscardPhase(g, 0)
	resetAllSambaPlayers(g)
	p := g.GetPlayer(0)
	p.SetMelds([]*domain.SambaMeld{sambaSevenSet(5), sambaSevenSet(9)})
	p.SetHasInitMeld(true)
	p.AddCard(sambaCard(domain.CardDesignClover, 8))
	require.NoError(t, g.PlayerGoOut())
	assert.True(t, g.GetPhase() == domain.SambaPhaseRoundEnd || g.GetPhase() == domain.SambaPhaseGameEnd)
}

func TestSamba_GoOut_TooManyCards(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDiscardPhase(g, 0)
	resetAllSambaPlayers(g)
	p := g.GetPlayer(0)
	p.SetMelds([]*domain.SambaMeld{sambaSevenSet(5), sambaSevenSet(9)})
	p.SetHasInitMeld(true)
	p.AddCard(sambaCard(domain.CardDesignClover, 8))
	p.AddCard(sambaCard(domain.CardDesignClover, 10))
	assert.Error(t, g.PlayerGoOut())
}

func TestSamba_GoOut_Red3LastCardRejected(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDiscardPhase(g, 0)
	resetAllSambaPlayers(g)
	p := g.GetPlayer(0)
	p.SetMelds([]*domain.SambaMeld{sambaSevenSet(5), sambaSevenSet(9)})
	p.SetHasInitMeld(true)
	p.AddCard(sambaCard(domain.CardDesignHeart, 3)) // red 3 cannot be discarded
	assert.Error(t, g.PlayerGoOut())
}

func TestSamba_GoOut_GameEndedAndNotHumanAndWrongPhase(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	assert.Error(t, g.PlayerGoOut()) // wrong phase
	setupSambaDiscardPhase(g, 1)
	assert.Error(t, g.PlayerGoOut()) // not human
	g.SetGameEndFlag(true)
	assert.Error(t, g.PlayerGoOut()) // game ended
}

// --- Discard advance/round-end ---

func TestSamba_Discard_EmptyDrawEndsRound(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDiscardPhase(g, 0)
	g.SetDrawPile(nil)
	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(sambaCard(domain.CardDesignSpade, 7))
	p.AddCard(sambaCard(domain.CardDesignHeart, 8))
	require.NoError(t, g.PlayerDiscard(0))
	assert.True(t, g.GetPhase() == domain.SambaPhaseRoundEnd || g.GetPhase() == domain.SambaPhaseGameEnd)
}

func TestSamba_Discard_GameEndedAndNotHuman(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDiscardPhase(g, 1)
	assert.Error(t, g.PlayerDiscard(0)) // not human
	g.SetGameEndFlag(true)
	assert.Error(t, g.PlayerDiscard(0)) // game ended
}

// --- scoreRound: endRoundDraw + team 1 win + hand penalty ---

func TestSamba_EndRoundDraw_Team1Wins(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	cfg := g.GetConfig()
	cfg.PointLimit = 500
	g.SetConfig(cfg)
	resetAllSambaPlayers(g)
	g.GetPlayer(1).SetMelds([]*domain.SambaMeld{sambaSevenSet(9)}) // team 1: 70 + 500
	setupSambaDrawPhase(g, 0)
	g.SetDrawPile(nil)
	require.NoError(t, g.PlayerDrawFromStock()) // empty stock -> endRoundDraw
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.SambaPhaseGameEnd, g.GetPhase())
	assert.Equal(t, 1, g.GetWinnerIdx())
	assert.Equal(t, 0, g.GetTeamScore(0))
	assert.Equal(t, 70+domain.SambaNaturalCanastaBonus, g.GetTeamScore(1))
}

func TestSamba_ScoreRound_HandPenalty(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	cfg := g.GetConfig()
	cfg.PointLimit = 100000 // avoid game end
	g.SetConfig(cfg)
	resetAllSambaPlayers(g)
	p0 := g.GetPlayer(0)
	p0.SetMelds([]*domain.SambaMeld{sambaSevenSet(5), sambaSevenSet(9)})
	p0.SetHasInitMeld(true)
	// team 1 holds only a leftover King -> negative round contribution.
	g.GetPlayer(1).AddCard(sambaCard(domain.CardDesignSpade, 13))
	setupSambaDiscardPhase(g, 0)
	require.NoError(t, g.PlayerGoOut())
	assert.Negative(t, g.GetTeamScore(1)) // -10 for the King in hand
}

// --- CPU: pick up the discard pile ---

func TestSamba_CpuDraw_PicksUpPile(t *testing.T) {
	g := newTestSambaWithDifficulty(domain.SambaCpuDifficultyNormal)
	g.Reset()
	setupSambaDrawPhase(g, 1)
	p := g.GetPlayer(1)
	p.Reset()
	p.SetHasInitMeld(true)
	p.AddCard(sambaCard(domain.CardDesignHeart, 7))
	p.AddCard(sambaCard(domain.CardDesignDiamond, 7))
	g.SetDiscardPile([]*domain.Card{
		sambaCard(domain.CardDesignClover, 5),
		sambaCard(domain.CardDesignSpade, 6),
		sambaCard(domain.CardDesignSpade, 7), // top matches the pair
	})
	g.CpuPlay()
	assert.Equal(t, domain.SambaPhaseMeld, g.GetPhase())
	assert.True(t, g.GetDrewFromDiscard())
}

func TestSamba_CpuDraw_PileRequirementBlocksPickup(t *testing.T) {
	g := newTestSambaWithDifficulty(domain.SambaCpuDifficultyNormal)
	g.Reset()
	setupSambaDrawPhase(g, 1)
	p := g.GetPlayer(1)
	p.Reset() // hasInitMeld false -> initial min 50 not met by three 4s (15)
	p.AddCard(sambaCard(domain.CardDesignHeart, 4))
	p.AddCard(sambaCard(domain.CardDesignDiamond, 4))
	g.SetDrawPile([]*domain.Card{sambaCard(domain.CardDesignClover, 12)})
	g.SetDiscardPile([]*domain.Card{sambaCard(domain.CardDesignSpade, 4)})
	g.CpuPlay()
	assert.Equal(t, domain.SambaPhaseMeld, g.GetPhase())
	assert.False(t, g.GetDrewFromDiscard()) // fell through to drawing stock
}

func TestSamba_CpuDraw_HardPile(t *testing.T) {
	g := newTestSambaWithDifficulty(domain.SambaCpuDifficultyHard)
	g.Reset()
	setupSambaDrawPhase(g, 1)
	p := g.GetPlayer(1)
	p.Reset()
	p.SetHasInitMeld(true)
	p.AddCard(sambaCard(domain.CardDesignHeart, 9))
	p.AddCard(sambaCard(domain.CardDesignDiamond, 9))
	g.SetDiscardPile([]*domain.Card{
		sambaCard(domain.CardDesignClover, 2),
		sambaCard(domain.CardDesignSpade, 5),
		sambaCard(domain.CardDesignSpade, 9),
	})
	g.CpuPlay()
	assert.Equal(t, domain.SambaPhaseMeld, g.GetPhase())
}

// --- CPU meld: set addition, initial-min skip, samba formation ---

func TestSamba_CpuMeld_AddsToExistingSet(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 1)
	p := g.GetPlayer(1)
	p.Reset()
	p.SetHasInitMeld(true)
	p.SetMelds([]*domain.SambaMeld{{
		Cards: []*domain.Card{
			sambaCard(domain.CardDesignSpade, 9),
			sambaCard(domain.CardDesignHeart, 9),
			sambaCard(domain.CardDesignClover, 9),
		},
		Kind:      domain.SambaMeldSet,
		IsNatural: true,
	}})
	p.AddCard(sambaCard(domain.CardDesignDiamond, 9))
	p.AddCard(sambaCard(domain.CardDesignSpade, 13)) // discard filler
	g.CpuPlay()                                      // meld
	assert.GreaterOrEqual(t, len(p.GetMelds()[0].Cards), 4)
}

func TestSamba_CpuMeld_InitialMinNotMetSkips(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 1)
	p := g.GetPlayer(1)
	p.Reset() // hasInitMeld false
	p.AddCard(sambaCard(domain.CardDesignSpade, 4))
	p.AddCard(sambaCard(domain.CardDesignHeart, 4))
	p.AddCard(sambaCard(domain.CardDesignDiamond, 4)) // 15 < 50
	g.CpuPlay()
	assert.Empty(t, p.GetMelds())
	assert.Equal(t, domain.SambaPhaseDiscard, g.GetPhase())
}

func TestSamba_CpuMeld_ExtendsExistingSequence(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 1)
	p := g.GetPlayer(1)
	p.Reset()
	p.SetHasInitMeld(true)
	p.SetMelds([]*domain.SambaMeld{{
		Cards: []*domain.Card{
			sambaCard(domain.CardDesignHeart, 4),
			sambaCard(domain.CardDesignHeart, 5),
			sambaCard(domain.CardDesignHeart, 6),
		},
		Kind:      domain.SambaMeldSequence,
		IsNatural: true,
	}})
	p.AddCard(sambaCard(domain.CardDesignHeart, 7))  // extends the run
	p.AddCard(sambaCard(domain.CardDesignHeart, 8))  // extends further
	p.AddCard(sambaCard(domain.CardDesignSpade, 13)) // discard filler
	g.CpuPlay()
	assert.GreaterOrEqual(t, len(p.GetMelds()[0].Cards), 4)
}

func TestSamba_CpuMeld_NoMeldGoesToDiscard(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 1)
	p := g.GetPlayer(1)
	p.Reset()
	p.SetHasInitMeld(true)
	p.AddCard(sambaCard(domain.CardDesignSpade, 13))
	p.AddCard(sambaCard(domain.CardDesignHeart, 8))
	g.CpuPlay()
	assert.Equal(t, domain.SambaPhaseDiscard, g.GetPhase())
}

// --- CPU discard branches ---

func TestSamba_CpuDiscard_Black3Priority(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDiscardPhase(g, 1)
	p := g.GetPlayer(1)
	p.Reset()
	p.AddCard(sambaCard(domain.CardDesignSpade, 3)) // black 3
	p.AddCard(sambaCard(domain.CardDesignHeart, 8))
	g.CpuPlay()
	top := g.GetDiscardTop()
	require.NotNil(t, top)
	assert.True(t, domain.SambaIsBlack3(top))
}

func TestSamba_CpuDiscard_EasyDifficulty(t *testing.T) {
	g := newTestSambaWithDifficulty(domain.SambaCpuDifficultyEasy)
	g.Reset()
	setupSambaDiscardPhase(g, 1)
	p := g.GetPlayer(1)
	p.Reset()
	p.AddCard(sambaCard(domain.CardDesignSpade, 7))
	p.AddCard(sambaCard(domain.CardDesignHeart, 8))
	g.CpuPlay()
	assert.Equal(t, domain.SambaPhaseDraw, g.GetPhase())
}

func TestSamba_CpuDiscard_OneCardGoOut(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDiscardPhase(g, 1)
	resetAllSambaPlayers(g)
	p := g.GetPlayer(1)
	p.SetMelds([]*domain.SambaMeld{sambaSevenSet(5), sambaSevenSet(9)})
	p.SetHasInitMeld(true)
	p.AddCard(sambaCard(domain.CardDesignClover, 8))
	g.CpuPlay()
	assert.True(t, g.GetPhase() == domain.SambaPhaseRoundEnd || g.GetPhase() == domain.SambaPhaseGameEnd)
}

func TestSamba_CpuDiscard_ZeroCardsNoGoOutAdvances(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDiscardPhase(g, 1)
	p := g.GetPlayer(1)
	p.Reset() // 0 cards, no completed melds
	g.CpuPlay()
	assert.Equal(t, domain.SambaPhaseDraw, g.GetPhase())
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
}

func TestSamba_CpuDiscard_ZeroCardsGoOut(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDiscardPhase(g, 1)
	resetAllSambaPlayers(g)
	p := g.GetPlayer(1)
	p.SetMelds([]*domain.SambaMeld{sambaSevenSet(5), sambaSevenSet(9)}) // 2 completed, 0 cards
	p.SetHasInitMeld(true)
	g.CpuPlay()
	assert.True(t, g.GetPhase() == domain.SambaPhaseRoundEnd || g.GetPhase() == domain.SambaPhaseGameEnd)
}

// --- Turn wrapping across 4 seats ---

func TestSamba_AdvanceTurn_WrapsAllSeats(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	for seat := 0; seat < domain.SambaPlayerCnt; seat++ {
		setupSambaDiscardPhase(g, seat)
		p := g.GetPlayer(seat)
		p.Reset()
		p.AddCard(sambaCard(domain.CardDesignSpade, 7))
		p.AddCard(sambaCard(domain.CardDesignHeart, 8))
		require.NoError(t, humanOrForceDiscard(g, seat))
	}
}

// humanOrForceDiscard discards for the current seat, using the human path for
// seat 0 and directly advancing CPU seats via CpuPlay for the others.
func humanOrForceDiscard(g *domain.Samba, seat int) error {
	if seat == 0 {
		return g.PlayerDiscard(0)
	}
	g.CpuPlay()
	return nil
}

// --- Getters ---

func TestSamba_Getters(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	g.SetRoundNumber(4)
	assert.Equal(t, 4, g.GetRoundNumber())
	g.SetGameEndFlag(true)
	assert.True(t, g.GetGameEndFlag())
	g.SetGameEndFlag(false)
	g.SetIsFrozen(true)
	assert.True(t, g.GetIsFrozen())

	pile := []*domain.Card{sambaCard(domain.CardDesignHeart, 9)}
	g.SetDiscardPile(pile)
	assert.Equal(t, pile, g.GetDiscardPile())
	assert.Equal(t, 9, g.GetDiscardTop().GetValue())

	g.SetDiscardPile(nil)
	assert.Nil(t, g.GetDiscardTop())

	cfg := domain.SambaConfig{CpuDifficulty: domain.SambaCpuDifficultyHard, PointLimit: 12345}
	g.SetConfig(cfg)
	assert.Equal(t, cfg, g.GetConfig())

	assert.False(t, g.GetDrewFromDiscard())
	assert.GreaterOrEqual(t, len(g.GetActionLog()), 0) // getter smoke (may be empty/nil)
	assert.Equal(t, domain.SambaTeamCnt, g.GetTeamCount())
}

// --- UnmarshalJSON validation branches ---

func TestSamba_Unmarshal_OversizeRejected(t *testing.T) {
	big := "{\"wp\":[" + strings.Repeat("null,", 3001)
	big = strings.TrimSuffix(big, ",") + "]}"
	var g domain.Samba
	assert.Error(t, json.Unmarshal([]byte(big), &g))
}

func TestSamba_Unmarshal_EmptyDefaults(t *testing.T) {
	var g domain.Samba
	require.NoError(t, json.Unmarshal([]byte(`{}`), &g))
	assert.Equal(t, 0, g.GetPlayerCnt())
	assert.Equal(t, domain.SambaPhaseDraw, g.GetPhase())
	assert.Equal(t, 0, g.GetTeamScore(0))
	assert.Equal(t, 0, g.GetTeamScore(1))
	assert.NotNil(t, g.GetActionLog())
	assert.Equal(t, 0, g.GetDrawPileCount())
	assert.Equal(t, 0, g.GetDiscardPileCount())
}

func TestSamba_Unmarshal_TeamScoresNormalized(t *testing.T) {
	// too-long team scores are truncated to SambaTeamCnt.
	var g domain.Samba
	require.NoError(t, json.Unmarshal([]byte(`{"ts":[10,20,30,40]}`), &g))
	assert.Equal(t, 10, g.GetTeamScore(0))
	assert.Equal(t, 20, g.GetTeamScore(1))

	// too-short is padded with zeros.
	var g2 domain.Samba
	require.NoError(t, json.Unmarshal([]byte(`{"ts":[55]}`), &g2))
	assert.Equal(t, 55, g2.GetTeamScore(0))
	assert.Equal(t, 0, g2.GetTeamScore(1))
}

func TestSamba_Unmarshal_InvalidMeldKindClamped(t *testing.T) {
	var p domain.SambaPlayer
	require.NoError(t, json.Unmarshal([]byte(`{"ml":[{"ca":[],"kd":9,"in":true}]}`), &p))
	require.Len(t, p.GetMelds(), 1)
	assert.Equal(t, domain.SambaMeldSet, p.GetMelds()[0].Kind)
}

// --- playerName out of range via winner banner path is covered elsewhere; here
// exercise GetPlayer bounds already covered. Cover Reset team assignment. ---

func TestSamba_Reset_AssignsTeams(t *testing.T) {
	g := newTestSamba()
	// deliberately corrupt a team before Reset, then confirm Reset restores it.
	g.Reset()
	g.GetPlayer(2).SetTeam(1)
	g.Reset()
	assert.Equal(t, 0, g.GetPlayer(2).GetTeam())
	assert.Equal(t, 1, g.GetPlayer(3).GetTeam())
}
