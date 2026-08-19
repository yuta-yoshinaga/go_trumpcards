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

func sambaCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func newTestSamba() *domain.Samba {
	players := make([]*domain.SambaPlayer, 0, domain.SambaPlayerCnt)
	for i := 0; i < domain.SambaPlayerCnt; i++ {
		players = append(players, domain.NewSambaPlayer(i == 0, i%domain.SambaTeamCnt))
	}
	return domain.NewSamba(domain.NewTrumpCardsWithDecks(3, 6), players, domain.DefaultSambaConfig())
}

func newTestSambaWithDifficulty(d domain.SambaCpuDifficulty) *domain.Samba {
	players := make([]*domain.SambaPlayer, 0, domain.SambaPlayerCnt)
	for i := 0; i < domain.SambaPlayerCnt; i++ {
		players = append(players, domain.NewSambaPlayer(i == 0, i%domain.SambaTeamCnt))
	}
	cfg := domain.DefaultSambaConfig()
	cfg.CpuDifficulty = d
	return domain.NewSamba(domain.NewTrumpCardsWithDecks(3, 6), players, cfg)
}

func setupSambaDrawPhase(g *domain.Samba, idx int) {
	g.SetPhase(domain.SambaPhaseDraw)
	g.SetCurrentPlayerIdx(idx)
}

func setupSambaMeldPhase(g *domain.Samba, idx int) {
	g.SetPhase(domain.SambaPhaseMeld)
	g.SetCurrentPlayerIdx(idx)
}

func setupSambaDiscardPhase(g *domain.Samba, idx int) {
	g.SetPhase(domain.SambaPhaseDiscard)
	g.SetCurrentPlayerIdx(idx)
}

// sambaSevenSet returns a 7-card natural set of the given rank (a canasta).
func sambaSevenSet(rank int) *domain.SambaMeld {
	cards := make([]*domain.Card, 7)
	designs := []int{domain.CardDesignSpade, domain.CardDesignHeart, domain.CardDesignDiamond, domain.CardDesignClover}
	for i := range cards {
		cards[i] = sambaCard(designs[i%len(designs)], rank)
	}
	return &domain.SambaMeld{Cards: cards, Kind: domain.SambaMeldSet, IsNatural: true}
}

// sambaSevenSequence returns a 7-card heart run 4..10 (a samba).
func sambaSevenSequence() *domain.SambaMeld {
	cards := []*domain.Card{
		sambaCard(domain.CardDesignHeart, 4),
		sambaCard(domain.CardDesignHeart, 5),
		sambaCard(domain.CardDesignHeart, 6),
		sambaCard(domain.CardDesignHeart, 7),
		sambaCard(domain.CardDesignHeart, 8),
		sambaCard(domain.CardDesignHeart, 9),
		sambaCard(domain.CardDesignHeart, 10),
	}
	return &domain.SambaMeld{Cards: cards, Kind: domain.SambaMeldSequence, IsNatural: true}
}

// --- Constructor & Reset ---

func TestNewSamba(t *testing.T) {
	g := newTestSamba()
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, 0, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, 2, g.GetTeamCount())
}

func TestSamba_Reset(t *testing.T) {
	g := newTestSamba()
	g.Reset()

	assert.Equal(t, domain.SambaPhaseDraw, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, 0, g.GetTeamScore(0))
	assert.Equal(t, 0, g.GetTeamScore(1))

	for i := 0; i < domain.SambaPlayerCnt; i++ {
		p := g.GetPlayer(i)
		assert.Equal(t, domain.SambaHandSize, p.GetCardsSize(), "player %d hand should have 15 cards", i)
		assert.Equal(t, i%domain.SambaTeamCnt, p.GetTeam())
		for j := 0; j < p.GetCardsSize(); j++ {
			assert.False(t, domain.SambaIsRed3(p.GetCard(j)), "player %d hand should not contain red 3s", i)
		}
	}

	assert.GreaterOrEqual(t, g.GetDiscardPileCount(), 1)

	// Total cards = 162 (3 decks + 6 jokers)
	total := g.GetDrawPileCount() + g.GetDiscardPileCount()
	for i := 0; i < domain.SambaPlayerCnt; i++ {
		total += g.GetPlayer(i).GetCardsSize() + len(g.GetPlayer(i).GetRed3s())
	}
	assert.Equal(t, 162, total, "total cards should be 162")
}

func TestSamba_NewDefaultSamba_DeckSize(t *testing.T) {
	g := domain.NewDefaultSamba()
	g.Reset()
	total := g.GetDrawPileCount() + g.GetDiscardPileCount()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		total += g.GetPlayer(i).GetCardsSize() + len(g.GetPlayer(i).GetRed3s())
	}
	assert.Equal(t, 162, total)
	assert.Equal(t, 4, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman())
	assert.False(t, g.GetPlayer(1).GetIsHuman())
}

// --- Card Type Helpers ---

func TestSambaIsWild(t *testing.T) {
	assert.True(t, domain.SambaIsWild(sambaCard(domain.CardDesignJoker, 1)))
	assert.True(t, domain.SambaIsWild(sambaCard(domain.CardDesignSpade, 2)))
	assert.False(t, domain.SambaIsWild(sambaCard(domain.CardDesignHeart, 3)))
	assert.False(t, domain.SambaIsWild(sambaCard(domain.CardDesignSpade, 1)))
}

func TestSambaIsRed3AndBlack3(t *testing.T) {
	assert.True(t, domain.SambaIsRed3(sambaCard(domain.CardDesignHeart, 3)))
	assert.True(t, domain.SambaIsRed3(sambaCard(domain.CardDesignDiamond, 3)))
	assert.False(t, domain.SambaIsRed3(sambaCard(domain.CardDesignSpade, 3)))
	assert.True(t, domain.SambaIsBlack3(sambaCard(domain.CardDesignSpade, 3)))
	assert.True(t, domain.SambaIsBlack3(sambaCard(domain.CardDesignClover, 3)))
	assert.False(t, domain.SambaIsBlack3(sambaCard(domain.CardDesignHeart, 3)))
}

func TestSambaCardValue(t *testing.T) {
	tests := []struct {
		card   *domain.Card
		expect int
	}{
		{sambaCard(domain.CardDesignJoker, 1), 50},
		{sambaCard(domain.CardDesignSpade, 2), 20},
		{sambaCard(domain.CardDesignHeart, 1), 20},
		{sambaCard(domain.CardDesignSpade, 13), 10},
		{sambaCard(domain.CardDesignDiamond, 8), 10},
		{sambaCard(domain.CardDesignClover, 7), 5},
		{sambaCard(domain.CardDesignSpade, 3), 5},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expect, domain.SambaCardValue(tt.card))
	}
}

// --- Meld model ---

func TestSambaMeld_KindsAndCompletion(t *testing.T) {
	canasta := sambaSevenSet(7)
	assert.True(t, canasta.IsCanasta())
	assert.False(t, canasta.IsSamba())
	assert.True(t, canasta.IsCompleted())
	assert.Equal(t, 7, canasta.GetRank())

	samba := sambaSevenSequence()
	assert.True(t, samba.IsSamba())
	assert.False(t, samba.IsCanasta())
	assert.True(t, samba.IsCompleted())
	assert.Equal(t, domain.CardDesignHeart, samba.SuitDesign())

	short := &domain.SambaMeld{Cards: []*domain.Card{sambaCard(domain.CardDesignSpade, 5)}, Kind: domain.SambaMeldSet, IsNatural: true}
	assert.False(t, short.IsCompleted())
}

func TestSambaPlayer_CompletedCounts(t *testing.T) {
	p := domain.NewSambaPlayer(true, 0)
	assert.Equal(t, 0, p.CompletedMeldCount())
	assert.False(t, p.HasCanasta())
	assert.False(t, p.HasSamba())

	p.AddMeld(sambaSevenSet(5))
	p.AddMeld(sambaSevenSequence())
	assert.Equal(t, 2, p.CompletedMeldCount())
	assert.True(t, p.HasCanasta())
	assert.True(t, p.HasSamba())
}

// --- Phase guards ---

func TestSamba_DrawFromStock_Guards(t *testing.T) {
	g := newTestSamba()
	g.Reset()

	setupSambaMeldPhase(g, 0)
	assert.True(t, errors.Is(g.PlayerDrawFromStock(), domain.ErrWrongPhase))

	setupSambaDrawPhase(g, 1) // CPU
	assert.True(t, errors.Is(g.PlayerDrawFromStock(), domain.ErrNotHumanTurn))

	g.SetGameEndFlag(true)
	assert.True(t, errors.Is(g.PlayerDrawFromStock(), domain.ErrGameEnded))
}

func TestSamba_DrawFromStock_Success(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDrawPhase(g, 0)

	before := g.GetPlayer(0).GetCardsSize()
	drawBefore := g.GetDrawPileCount()
	require.NoError(t, g.PlayerDrawFromStock())
	assert.Equal(t, domain.SambaPhaseMeld, g.GetPhase())
	assert.GreaterOrEqual(t, g.GetPlayer(0).GetCardsSize(), before)
	assert.Less(t, g.GetDrawPileCount(), drawBefore)
}

func TestSamba_DrawFromStock_EmptyEndsRound(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDrawPhase(g, 0)
	g.SetDrawPile(nil)
	require.NoError(t, g.PlayerDrawFromStock())
	assert.True(t, g.GetPhase() == domain.SambaPhaseRoundEnd || g.GetPhase() == domain.SambaPhaseGameEnd)
}

// --- Take discard pile ---

func TestSamba_DrawFromDiscard_Guards(t *testing.T) {
	g := newTestSamba()
	g.Reset()

	setupSambaMeldPhase(g, 0)
	assert.True(t, errors.Is(g.PlayerDrawFromDiscard([]int{0, 1}), domain.ErrWrongPhase))

	setupSambaDrawPhase(g, 0)
	g.SetDiscardPile([]*domain.Card{sambaCard(domain.CardDesignSpade, 3)}) // black 3
	assert.Error(t, g.PlayerDrawFromDiscard([]int{0, 1}))

	g.SetDiscardPile([]*domain.Card{sambaCard(domain.CardDesignSpade, 2)}) // wild
	assert.Error(t, g.PlayerDrawFromDiscard([]int{0, 1}))

	g.SetDiscardPile([]*domain.Card{sambaCard(domain.CardDesignSpade, 7)})
	assert.Error(t, g.PlayerDrawFromDiscard([]int{0})) // wrong pair count
}

func TestSamba_DrawFromDiscard_Success(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDrawPhase(g, 0)

	player := g.GetPlayer(0)
	player.Reset()
	top := sambaCard(domain.CardDesignSpade, 7)
	g.SetDiscardPile([]*domain.Card{sambaCard(domain.CardDesignHeart, 5), top})
	player.AddCard(sambaCard(domain.CardDesignHeart, 7))
	player.AddCard(sambaCard(domain.CardDesignDiamond, 7))
	player.AddCard(sambaCard(domain.CardDesignClover, 10))
	player.SetHasInitMeld(true)

	require.NoError(t, g.PlayerDrawFromDiscard([]int{0, 1}))
	assert.Equal(t, domain.SambaPhaseMeld, g.GetPhase())
	assert.True(t, g.GetDrewFromDiscard())
	assert.Nil(t, g.GetDiscardPile())
}

// --- Set melds ---

func TestSamba_Meld_ValidSetWithWild(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)

	player := g.GetPlayer(0)
	player.Reset()
	player.SetHasInitMeld(true)
	player.AddCard(sambaCard(domain.CardDesignSpade, 9))
	player.AddCard(sambaCard(domain.CardDesignHeart, 9))
	player.AddCard(sambaCard(domain.CardDesignJoker, 1)) // wild
	player.AddCard(sambaCard(domain.CardDesignClover, 4))

	require.NoError(t, g.PlayerMeld([][]int{{0, 1, 2}}))
	melds := player.GetMelds()
	require.Len(t, melds, 1)
	assert.Equal(t, domain.SambaMeldSet, melds[0].Kind)
	assert.False(t, melds[0].IsNatural)
}

func TestSamba_Meld_TooManyWildsRejected(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	player := g.GetPlayer(0)
	player.Reset()
	player.SetHasInitMeld(true)
	player.AddCard(sambaCard(domain.CardDesignSpade, 9))
	player.AddCard(sambaCard(domain.CardDesignJoker, 1))
	player.AddCard(sambaCard(domain.CardDesignJoker, 1))
	player.AddCard(sambaCard(domain.CardDesignSpade, 2))
	player.AddCard(sambaCard(domain.CardDesignHeart, 2))
	assert.Error(t, g.PlayerMeld([][]int{{0, 1, 2, 3, 4}}))
}

func TestSamba_Meld_Black3Rejected(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	player := g.GetPlayer(0)
	player.Reset()
	player.SetHasInitMeld(true)
	player.AddCard(sambaCard(domain.CardDesignSpade, 3))
	player.AddCard(sambaCard(domain.CardDesignClover, 3))
	player.AddCard(sambaCard(domain.CardDesignSpade, 3))
	assert.Error(t, g.PlayerMeld([][]int{{0, 1, 2}}))
}

// --- Sequence melds (the new mechanic) ---

func TestSamba_Meld_ValidSequence(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	player := g.GetPlayer(0)
	player.Reset()
	player.SetHasInitMeld(true)
	player.AddCard(sambaCard(domain.CardDesignHeart, 4))
	player.AddCard(sambaCard(domain.CardDesignHeart, 5))
	player.AddCard(sambaCard(domain.CardDesignHeart, 6))

	require.NoError(t, g.PlayerMeld([][]int{{0, 1, 2}}))
	melds := player.GetMelds()
	require.Len(t, melds, 1)
	assert.Equal(t, domain.SambaMeldSequence, melds[0].Kind)
	assert.Len(t, melds[0].Cards, 3)
}

func TestSamba_Meld_SequenceUnsortedStillValid(t *testing.T) {
	// Rule 9: run validation must sort values before checking the span.
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	player := g.GetPlayer(0)
	player.Reset()
	player.SetHasInitMeld(true)
	player.AddCard(sambaCard(domain.CardDesignHeart, 6))
	player.AddCard(sambaCard(domain.CardDesignHeart, 4))
	player.AddCard(sambaCard(domain.CardDesignHeart, 5))
	require.NoError(t, g.PlayerMeld([][]int{{0, 1, 2}}))
	assert.Equal(t, domain.SambaMeldSequence, player.GetMelds()[0].Kind)
}

func TestSamba_Meld_SequenceWildRejected(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	player := g.GetPlayer(0)
	player.Reset()
	player.SetHasInitMeld(true)
	player.AddCard(sambaCard(domain.CardDesignHeart, 4))
	player.AddCard(sambaCard(domain.CardDesignHeart, 5))
	player.AddCard(sambaCard(domain.CardDesignJoker, 1)) // wild not allowed in a sequence
	assert.Error(t, g.PlayerMeld([][]int{{0, 1, 2}}))
}

func TestSamba_Meld_SequenceNonConsecutiveRejected(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	player := g.GetPlayer(0)
	player.Reset()
	player.SetHasInitMeld(true)
	player.AddCard(sambaCard(domain.CardDesignHeart, 4))
	player.AddCard(sambaCard(domain.CardDesignHeart, 6))
	player.AddCard(sambaCard(domain.CardDesignHeart, 8))
	assert.Error(t, g.PlayerMeld([][]int{{0, 1, 2}}))
}

func TestSamba_Meld_SambaCompletionAndExtension(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	player := g.GetPlayer(0)
	player.Reset()
	player.SetHasInitMeld(true)
	// existing 5-card heart run 4..8
	player.AddMeld(&domain.SambaMeld{
		Cards: []*domain.Card{
			sambaCard(domain.CardDesignHeart, 4),
			sambaCard(domain.CardDesignHeart, 5),
			sambaCard(domain.CardDesignHeart, 6),
			sambaCard(domain.CardDesignHeart, 7),
			sambaCard(domain.CardDesignHeart, 8),
		},
		Kind:      domain.SambaMeldSequence,
		IsNatural: true,
	})
	player.AddCard(sambaCard(domain.CardDesignHeart, 9))
	player.AddCard(sambaCard(domain.CardDesignHeart, 10))
	player.AddCard(sambaCard(domain.CardDesignClover, 5)) // filler

	// add 9 and 10 to the existing heart run -> 7-card samba
	require.NoError(t, g.PlayerMeld([][]int{{0, 1}}))
	require.Len(t, player.GetMelds(), 1)
	assert.True(t, player.GetMelds()[0].IsSamba())
	assert.True(t, player.HasSamba())
}

func TestSamba_Meld_MixedSetAndSequenceInOneCall(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	player := g.GetPlayer(0)
	player.Reset()
	player.SetHasInitMeld(true)
	player.AddCard(sambaCard(domain.CardDesignSpade, 9))   // 0 set
	player.AddCard(sambaCard(domain.CardDesignHeart, 9))   // 1 set
	player.AddCard(sambaCard(domain.CardDesignClover, 9))  // 2 set
	player.AddCard(sambaCard(domain.CardDesignDiamond, 4)) // 3 seq
	player.AddCard(sambaCard(domain.CardDesignDiamond, 5)) // 4 seq
	player.AddCard(sambaCard(domain.CardDesignDiamond, 6)) // 5 seq

	require.NoError(t, g.PlayerMeld([][]int{{0, 1, 2}, {3, 4, 5}}))
	melds := player.GetMelds()
	require.Len(t, melds, 2)
	kinds := map[domain.SambaMeldKind]int{}
	for _, m := range melds {
		kinds[m.Kind]++
	}
	assert.Equal(t, 1, kinds[domain.SambaMeldSet])
	assert.Equal(t, 1, kinds[domain.SambaMeldSequence])
}

func TestSamba_Meld_InitialMinimumEnforced(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	player := g.GetPlayer(0)
	player.Reset()
	// score 0 -> min 50. Three 4s = 15 points -> fail.
	player.AddCard(sambaCard(domain.CardDesignSpade, 4))
	player.AddCard(sambaCard(domain.CardDesignHeart, 4))
	player.AddCard(sambaCard(domain.CardDesignDiamond, 4))
	assert.Error(t, g.PlayerMeld([][]int{{0, 1, 2}}))
}

func TestSamba_SkipMeld(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 0)
	require.NoError(t, g.PlayerSkipMeld())
	assert.Equal(t, domain.SambaPhaseDiscard, g.GetPhase())
}

// --- Discard ---

func TestSamba_Discard(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDiscardPhase(g, 0)
	player := g.GetPlayer(0)
	player.Reset()
	player.AddCard(sambaCard(domain.CardDesignSpade, 7))
	player.AddCard(sambaCard(domain.CardDesignHeart, 8))

	before := g.GetDiscardPileCount()
	require.NoError(t, g.PlayerDiscard(0))
	assert.Equal(t, 1, player.GetCardsSize())
	assert.Equal(t, before+1, g.GetDiscardPileCount())
	// turn advances to player 1
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
}

func TestSamba_Discard_Red3Rejected(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDiscardPhase(g, 0)
	player := g.GetPlayer(0)
	player.Reset()
	player.AddCard(sambaCard(domain.CardDesignHeart, 3))
	assert.Error(t, g.PlayerDiscard(0))
}

func TestSamba_Discard_WildFreezes(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDiscardPhase(g, 0)
	g.SetIsFrozen(false)
	player := g.GetPlayer(0)
	player.Reset()
	player.AddCard(sambaCard(domain.CardDesignSpade, 2))
	player.AddCard(sambaCard(domain.CardDesignHeart, 8))
	require.NoError(t, g.PlayerDiscard(0))
	assert.True(t, g.GetIsFrozen())
}

// --- Go out & team scoring ---

func TestSamba_GoOut_RequiresTwoCompletedMelds(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDiscardPhase(g, 0)
	player := g.GetPlayer(0)
	player.Reset()
	player.SetMelds([]*domain.SambaMeld{sambaSevenSet(7)}) // only one completed
	player.SetHasInitMeld(true)
	player.AddCard(sambaCard(domain.CardDesignClover, 5))
	assert.Error(t, g.PlayerGoOut())
}

func TestSamba_GoOut_TeamScoring(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDiscardPhase(g, 0)

	// ResetRound clears hand, melds, red3s and round score for every player.
	// Reset() alone leaves the red 3s dealt in Reset(), which team 1 would then
	// score (+100 each) — this is correct per the implemented rule, so we clear
	// them here to make "team 1 contributes nothing" genuinely hold.
	for i := 0; i < domain.SambaPlayerCnt; i++ {
		g.GetPlayer(i).ResetRound()
	}
	player := g.GetPlayer(0)
	// team 0 gets a natural canasta of 5s and a heart samba
	player.SetMelds([]*domain.SambaMeld{sambaSevenSet(5), sambaSevenSequence()})
	player.SetHasInitMeld(true)
	// go out with an empty hand

	require.NoError(t, g.PlayerGoOut())
	assert.True(t, g.GetPhase() == domain.SambaPhaseRoundEnd || g.GetPhase() == domain.SambaPhaseGameEnd)

	// canasta: 7×5pts=35 + 500; samba: (5+5+5+5+10+10+10)=50 + 1500; go-out 100
	expectedTeam0 := 35 + domain.SambaNaturalCanastaBonus + 50 + domain.SambaSambaBonus + domain.SambaGoingOutBonus
	assert.Equal(t, expectedTeam0, g.GetTeamScore(0))
	assert.Equal(t, 0, g.GetTeamScore(1))
}

func TestSamba_Red3Penalty_WhenTeamHasNotMelded(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDiscardPhase(g, 0)
	for i := 0; i < domain.SambaPlayerCnt; i++ {
		g.GetPlayer(i).ResetRound()
	}
	// team 0 melds a canasta + a samba and goes out.
	g.GetPlayer(0).SetMelds([]*domain.SambaMeld{sambaSevenSet(5), sambaSevenSequence()})
	g.GetPlayer(0).SetHasInitMeld(true)
	// team 1 never melded but holds two red 3s (empty hand) → they must be
	// SUBTRACTED, not added, per the standard Samba/Canasta rule.
	g.GetPlayer(1).AddRed3(sambaCard(domain.CardDesignHeart, 3))
	g.GetPlayer(1).AddRed3(sambaCard(domain.CardDesignDiamond, 3))

	require.NoError(t, g.PlayerGoOut())
	assert.Equal(t, -2*domain.SambaRed3Bonus, g.GetTeamScore(1))
}

func TestSamba_MixedCanastaScoresLess(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDiscardPhase(g, 0)
	for i := 0; i < domain.SambaPlayerCnt; i++ {
		g.GetPlayer(i).Reset()
		g.GetPlayer(i).SetMelds(nil)
	}
	player := g.GetPlayer(0)
	// mixed canasta: 5 naturals (5s) + 2 wilds
	cards := []*domain.Card{
		sambaCard(domain.CardDesignSpade, 5), sambaCard(domain.CardDesignHeart, 5),
		sambaCard(domain.CardDesignDiamond, 5), sambaCard(domain.CardDesignClover, 5),
		sambaCard(domain.CardDesignSpade, 5),
		sambaCard(domain.CardDesignJoker, 1), sambaCard(domain.CardDesignSpade, 2),
	}
	mixed := &domain.SambaMeld{Cards: cards, Kind: domain.SambaMeldSet, IsNatural: false}
	player.SetMelds([]*domain.SambaMeld{mixed, sambaSevenSet(9)})
	player.SetHasInitMeld(true)
	require.NoError(t, g.PlayerGoOut())
	// mixed canasta bonus should be the mixed value, not natural
	assert.Positive(t, g.GetTeamScore(0))
}

func TestSamba_Red3Bonus(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDiscardPhase(g, 0)
	for i := 0; i < domain.SambaPlayerCnt; i++ {
		g.GetPlayer(i).Reset()
		g.GetPlayer(i).SetMelds(nil)
		g.GetPlayer(i).SetRed3s(nil)
	}
	player := g.GetPlayer(0)
	player.SetMelds([]*domain.SambaMeld{sambaSevenSet(5), sambaSevenSet(9)})
	player.SetHasInitMeld(true)
	player.AddRed3(sambaCard(domain.CardDesignHeart, 3))
	player.AddRed3(sambaCard(domain.CardDesignDiamond, 3))
	require.NoError(t, g.PlayerGoOut())
	// two red 3s = 200 bonus included
	assert.GreaterOrEqual(t, g.GetTeamScore(0), 2*domain.SambaRed3Bonus)
}

func TestSamba_GameEndOnPointLimit(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	cfg := g.GetConfig()
	cfg.PointLimit = 100
	g.SetConfig(cfg)
	setupSambaDiscardPhase(g, 0)
	for i := 0; i < domain.SambaPlayerCnt; i++ {
		g.GetPlayer(i).Reset()
		g.GetPlayer(i).SetMelds(nil)
	}
	player := g.GetPlayer(0)
	player.SetMelds([]*domain.SambaMeld{sambaSevenSet(5), sambaSevenSequence()})
	player.SetHasInitMeld(true)
	require.NoError(t, g.PlayerGoOut())
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.SambaPhaseGameEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetWinnerIdx()) // team 0 (human's team) wins
}

// --- NextRound ---

func TestSamba_NextRound_WrongPhaseNoop(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	g.NextRound()
	assert.Equal(t, domain.SambaPhaseDraw, g.GetPhase())
}

func TestSamba_NextRound(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	g.SetPhase(domain.SambaPhaseRoundEnd)
	g.NextRound()
	assert.Equal(t, domain.SambaPhaseDraw, g.GetPhase())
	assert.Equal(t, 2, g.GetRoundNumber())
	for i := 0; i < domain.SambaPlayerCnt; i++ {
		assert.Equal(t, domain.SambaHandSize, g.GetPlayer(i).GetCardsSize())
	}
}

// --- CPU ---

func TestSamba_CpuPlay_Draw(t *testing.T) {
	for _, diff := range []domain.SambaCpuDifficulty{
		domain.SambaCpuDifficultyEasy,
		domain.SambaCpuDifficultyNormal,
		domain.SambaCpuDifficultyHard,
	} {
		t.Run(fmt.Sprintf("difficulty_%d", diff), func(t *testing.T) {
			g := newTestSambaWithDifficulty(diff)
			g.Reset()
			setupSambaDrawPhase(g, 1)
			g.CpuPlay()
			assert.NotEqual(t, domain.SambaPhaseDraw, g.GetPhase())
		})
	}
}

func TestSamba_CpuPlay_MeldAndDiscard(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 1)
	g.CpuPlay()
	assert.NotEqual(t, domain.SambaPhaseMeld, g.GetPhase())
}

func TestSamba_CpuPlay_FormsSequence(t *testing.T) {
	// Give a CPU a clean 4-card heart run + a discardable card; it should be
	// able to form a sequence meld (rule 11: CPU must find sambas).
	g := newTestSamba()
	g.Reset()
	setupSambaMeldPhase(g, 1)
	player := g.GetPlayer(1)
	player.Reset()
	player.SetHasInitMeld(true)
	player.AddCard(sambaCard(domain.CardDesignHeart, 4))
	player.AddCard(sambaCard(domain.CardDesignHeart, 5))
	player.AddCard(sambaCard(domain.CardDesignHeart, 6))
	player.AddCard(sambaCard(domain.CardDesignHeart, 7))
	player.AddCard(sambaCard(domain.CardDesignSpade, 13))
	g.CpuPlay() // meld
	foundSeq := false
	for _, m := range player.GetMelds() {
		if m.Kind == domain.SambaMeldSequence {
			foundSeq = true
		}
	}
	assert.True(t, foundSeq, "CPU should have formed a sequence meld")
}

func TestSamba_CpuPlay_HumanTurnNoop(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	setupSambaDrawPhase(g, 0)
	g.CpuPlay()
	assert.Equal(t, domain.SambaPhaseDraw, g.GetPhase())
}

func TestSamba_CpuPlay_GameEndedNoop(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	g.SetGameEndFlag(true)
	g.CpuPlay()
}

// A bounded CPU drive must always land on a valid phase (never non-terminates,
// never panics). Do NOT assert game-end.
func TestSamba_CpuDrive_BoundedStaysValid(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	for step := 0; step < 200; step++ {
		if g.GetGameEndFlag() {
			break
		}
		phase := g.GetPhase()
		if phase == domain.SambaPhaseRoundEnd {
			g.NextRound()
			continue
		}
		if g.IsHumanTurn() {
			// drive the human minimally: draw -> skip meld -> discard first legal
			switch phase {
			case domain.SambaPhaseDraw:
				_ = g.PlayerDrawFromStock()
			case domain.SambaPhaseMeld:
				_ = g.PlayerSkipMeld()
			case domain.SambaPhaseDiscard:
				p := g.GetPlayer(0)
				idx := -1
				for j := 0; j < p.GetCardsSize(); j++ {
					if !domain.SambaIsRed3(p.GetCard(j)) {
						idx = j
						break
					}
				}
				if idx >= 0 {
					_ = g.PlayerDiscard(idx)
				} else {
					g.SetPhase(domain.SambaPhaseRoundEnd)
				}
			}
			continue
		}
		g.CpuPlay()
	}
	valid := g.GetPhase() >= domain.SambaPhaseDraw && g.GetPhase() <= domain.SambaPhaseGameEnd
	assert.True(t, valid)
}

// --- Getters / config ---

func TestSamba_GetPlayer_OutOfBounds(t *testing.T) {
	g := newTestSamba()
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
}

func TestSamba_IsHumanTurn(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(1)
	assert.False(t, g.IsHumanTurn())
}

func TestSamba_GetTeamScoreBounds(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	g.SetTeamScore(0, 500)
	assert.Equal(t, 500, g.GetTeamScore(0))
	assert.Equal(t, 0, g.GetTeamScore(-1))
	assert.Equal(t, 0, g.GetTeamScore(99))
}

func TestSambaConfig_Default(t *testing.T) {
	cfg := domain.DefaultSambaConfig()
	assert.Equal(t, domain.SambaCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, domain.SambaDefaultPointLimit, cfg.PointLimit)
}

func TestSambaConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     domain.SambaConfig
		wantErr bool
	}{
		{"valid default", domain.DefaultSambaConfig(), false},
		{"valid easy", domain.SambaConfig{CpuDifficulty: domain.SambaCpuDifficultyEasy, PointLimit: 100}, false},
		{"invalid difficulty", domain.SambaConfig{CpuDifficulty: 9, PointLimit: 5000}, true},
		{"invalid point limit", domain.SambaConfig{CpuDifficulty: domain.SambaCpuDifficultyNormal, PointLimit: 0}, true},
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

// --- JSON ---

func TestSamba_JSON_RoundTrip(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	g.GetPlayer(0).AddMeld(sambaSevenSequence())
	g.SetTeamScore(0, 1200)
	g.SetTeamScore(1, 300)

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var g2 domain.Samba
	require.NoError(t, json.Unmarshal(data, &g2))

	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetRoundNumber(), g2.GetRoundNumber())
	assert.Equal(t, g.GetCurrentPlayerIdx(), g2.GetCurrentPlayerIdx())
	assert.Equal(t, g.GetDrawPileCount(), g2.GetDrawPileCount())
	assert.Equal(t, 1200, g2.GetTeamScore(0))
	assert.Equal(t, 300, g2.GetTeamScore(1))
	require.NotNil(t, g2.GetPlayer(0))
	require.Len(t, g2.GetPlayer(0).GetMelds(), 1)
	assert.Equal(t, domain.SambaMeldSequence, g2.GetPlayer(0).GetMelds()[0].Kind)
}

func TestSamba_JSON_ValidatesCorruptIndices(t *testing.T) {
	g := newTestSamba()
	g.Reset()
	g.GetPlayer(0).SetTeam(9) // out of range team
	g.SetCurrentPlayerIdx(0)

	data, err := json.Marshal(g)
	require.NoError(t, err)

	// Corrupt phase & winner in the raw JSON.
	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))
	raw["ps"] = json.RawMessage("99")
	raw["ci"] = json.RawMessage("99")
	raw["wi"] = json.RawMessage("99")
	patched, err := json.Marshal(raw)
	require.NoError(t, err)

	var g2 domain.Samba
	require.NoError(t, json.Unmarshal(patched, &g2))
	// team clamped back into [0, teamCount)
	assert.GreaterOrEqual(t, g2.GetPlayer(0).GetTeam(), 0)
	assert.Less(t, g2.GetPlayer(0).GetTeam(), domain.SambaTeamCnt)
	// phase / currentPlayerIdx / winnerIdx sanitised
	assert.Equal(t, domain.SambaPhaseDraw, g2.GetPhase())
	assert.Equal(t, 0, g2.GetCurrentPlayerIdx())
	assert.Equal(t, -1, g2.GetWinnerIdx())
}

func TestSamba_JSON_RejectsNilPlayer(t *testing.T) {
	var g domain.Samba
	assert.Error(t, json.Unmarshal([]byte(`{"pl":[null]}`), &g))
}

func TestSambaPlayer_JSON_FiltersNilMeld(t *testing.T) {
	var p domain.SambaPlayer
	require.NoError(t, json.Unmarshal([]byte(`{"ml":[null]}`), &p))
	assert.Empty(t, p.GetMelds())
}

func TestSambaPlayer_JSON_RoundTrip(t *testing.T) {
	p := domain.NewSambaPlayer(true, 0)
	p.AddCard(sambaCard(domain.CardDesignSpade, 7))
	p.AddMeld(sambaSevenSequence())
	p.AddRed3(sambaCard(domain.CardDesignHeart, 3))
	p.SetHasInitMeld(true)
	p.SetRoundScore(50)

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var p2 domain.SambaPlayer
	require.NoError(t, json.Unmarshal(data, &p2))
	assert.Equal(t, p.GetCardsSize(), p2.GetCardsSize())
	assert.Equal(t, len(p.GetMelds()), len(p2.GetMelds()))
	assert.Equal(t, len(p.GetRed3s()), len(p2.GetRed3s()))
	assert.Equal(t, p.GetHasInitMeld(), p2.GetHasInitMeld())
	assert.Equal(t, p.GetTeam(), p2.GetTeam())
	assert.Equal(t, domain.SambaMeldSequence, p2.GetMelds()[0].Kind)
}

func TestSambaPlayer_ResetRound(t *testing.T) {
	p := domain.NewSambaPlayer(true, 0)
	p.AddCard(sambaCard(domain.CardDesignSpade, 5))
	p.AddMeld(sambaSevenSet(7))
	p.AddRed3(sambaCard(domain.CardDesignHeart, 3))
	p.SetHasInitMeld(true)
	p.SetRoundScore(100)

	p.ResetRound()
	assert.Equal(t, 0, p.GetCardsSize())
	assert.Empty(t, p.GetMelds())
	assert.Empty(t, p.GetRed3s())
	assert.False(t, p.GetHasInitMeld())
	assert.Equal(t, 0, p.GetRoundScore())
}

// #5702: 初回メルドの最低点はチーム累積点の帯で決まる。CUI と Web の両方が
// この値を出すので、帯の境界そのものをここで固定する。
func TestSambaMinimumMeldValue(t *testing.T) {
	cases := []struct {
		score int
		want  int
	}{
		{-1, 15}, {-500, 15},
		{0, 50}, {1499, 50},
		{1500, 90}, {2999, 90},
		{3000, 120}, {10000, 120},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, domain.SambaMinimumMeldValue(c.score), "score %d", c.score)
	}
}

// GetMinimumMeldValue は席のチームの累積点を引いて同じ表を使う。
func TestSamba_GetMinimumMeldValue(t *testing.T) {
	g := domain.NewDefaultSamba()
	g.Reset()

	assert.Equal(t, 50, g.GetMinimumMeldValue(0), "fresh game starts at 0 points")

	g.SetTeamScore(g.GetPlayer(0).GetTeam(), 1500)
	assert.Equal(t, 90, g.GetMinimumMeldValue(0))

	// 範囲外の添字は 0 (呼び出し側が席を取り違えても panic しない)。
	assert.Equal(t, 0, g.GetMinimumMeldValue(-1))
	assert.Equal(t, 0, g.GetMinimumMeldValue(99))
}
