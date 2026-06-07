package domain

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestYaniv builds a 4-player game with a deterministic RNG and the default
// config, then resets it so a round is in progress.
func newTestYaniv() *Yaniv {
	g := NewDefaultYaniv()
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()
	return g
}

// setYanivHand replaces a player's hand with the given cards.
func setYanivHand(p *YanivPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestNewYaniv_Defaults(t *testing.T) {
	g := NewDefaultYaniv()
	assert.Equal(t, YanivPlayerCnt, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman())
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Nil(t, g.GetPlayer(99))
}

func TestYaniv_ResetDealsHands(t *testing.T) {
	g := newTestYaniv()
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, YanivPhaseDiscard, g.GetPhase())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, YanivHandSize, g.GetPlayer(i).GetCardsSize())
		assert.Equal(t, 0, g.GetPlayer(i).GetScore())
	}
	assert.NotNil(t, g.GetDiscardTop())
	assert.Len(t, g.GetPickupCards(), 1)
	assert.Positive(t, g.GetDrawPileCount())
}

func TestYanivValidCombo(t *testing.T) {
	single := []*Card{NewCard(CardDesignSpade, 7, false)}
	pair := []*Card{NewCard(CardDesignSpade, 7, false), NewCard(CardDesignHeart, 7, false)}
	trio := []*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignClover, 7, false),
	}
	run := []*Card{
		NewCard(CardDesignSpade, 4, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 6, false),
	}
	runUnsorted := []*Card{
		NewCard(CardDesignSpade, 6, false),
		NewCard(CardDesignSpade, 4, false),
		NewCard(CardDesignSpade, 5, false),
	}
	badMixed := []*Card{NewCard(CardDesignSpade, 7, false), NewCard(CardDesignHeart, 8, false)}
	badRunShort := []*Card{NewCard(CardDesignSpade, 4, false), NewCard(CardDesignSpade, 5, false)}
	badRunSuit := []*Card{
		NewCard(CardDesignSpade, 4, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignSpade, 6, false),
	}
	jokerPair := []*Card{NewCard(CardDesignJoker, 1, false), NewCard(CardDesignJoker, 2, false)}

	assert.True(t, YanivValidCombo(single))
	assert.True(t, YanivValidCombo(pair))
	assert.True(t, YanivValidCombo(trio))
	assert.True(t, YanivValidCombo(run))
	assert.True(t, YanivValidCombo(runUnsorted))
	assert.False(t, YanivValidCombo(nil))
	assert.False(t, YanivValidCombo(badMixed))
	assert.False(t, YanivValidCombo(badRunShort))
	assert.False(t, YanivValidCombo(badRunSuit))
	assert.False(t, YanivValidCombo(jokerPair)) // jokers cannot form a value-set
}

func TestYaniv_HumanDiscardSingleThenDrawStock(t *testing.T) {
	g := newTestYaniv()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(YanivPhaseDiscard)
	setYanivHand(g.GetPlayer(0),
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	)
	g.SetPickupCards([]*Card{NewCard(CardDesignDiamond, 2, false)})
	g.SetDrawPile([]*Card{NewCard(CardDesignHeart, 8, false)})

	require.NoError(t, g.PlayerDiscard([]int{0})) // discard 9♠
	assert.Equal(t, YanivPhaseDraw, g.GetPhase())
	assert.Equal(t, 2, g.GetPlayer(0).GetCardsSize())

	require.NoError(t, g.PlayerDrawFromStock())
	// Turn advances away from the human; pickup becomes the card just discarded.
	assert.Equal(t, YanivPhaseDiscard, g.GetPhase())
	assert.NotEqual(t, 0, g.GetCurrentPlayerIdx())
	require.Len(t, g.GetPickupCards(), 1)
	assert.Equal(t, 9, g.GetPickupCards()[0].GetValue())
}

func TestYaniv_HumanDiscardComboPair(t *testing.T) {
	g := newTestYaniv()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(YanivPhaseDiscard)
	setYanivHand(g.GetPlayer(0),
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignHeart, 8, false),
		NewCard(CardDesignClover, 2, false),
	)
	g.SetDrawPile([]*Card{NewCard(CardDesignHeart, 5, false)})
	require.NoError(t, g.PlayerDiscard([]int{0, 1}))
	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
	assert.Len(t, g.GetPendingDiscard(), 2)
}

func TestYaniv_DiscardInvalidCombo(t *testing.T) {
	g := newTestYaniv()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(YanivPhaseDiscard)
	setYanivHand(g.GetPlayer(0),
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignHeart, 3, false),
	)
	assert.Error(t, g.PlayerDiscard([]int{0, 1})) // 8 and 3 are not a valid combo
	assert.Error(t, g.PlayerDiscard([]int{}))     // empty
	assert.Error(t, g.PlayerDiscard([]int{5}))    // out of range
}

func TestYaniv_DrawFromPickupEnds(t *testing.T) {
	g := newTestYaniv()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(YanivPhaseDraw)
	// Simulate the previous player having discarded a 3-card run; only ends are takeable.
	g.SetPickupCards([]*Card{
		NewCard(CardDesignSpade, 4, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 6, false),
	})
	g.pendingDiscard = []*Card{NewCard(CardDesignHeart, 9, false)}
	before := g.GetPlayer(0).GetCardsSize()
	require.NoError(t, g.PlayerDrawFromPickup(1)) // take the last (6♠)
	assert.Equal(t, before+1, g.GetPlayer(0).GetCardsSize())
	// The two unpicked cards are buried; the pending discard becomes the new pickup.
	require.Len(t, g.GetPickupCards(), 1)
	assert.Equal(t, 9, g.GetPickupCards()[0].GetValue())
}

func TestYaniv_DrawFromPickupEmptyErrors(t *testing.T) {
	g := newTestYaniv()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(YanivPhaseDraw)
	g.SetPickupCards([]*Card{})
	assert.Error(t, g.PlayerDrawFromPickup(0))
}

func TestYaniv_DeclareYanivSuccess(t *testing.T) {
	g := newTestYaniv()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(YanivPhaseDiscard)
	setYanivHand(g.GetPlayer(0), NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 2, false)) // 3
	setYanivHand(g.GetPlayer(1), NewCard(CardDesignSpade, 10, false))                                    // 10
	setYanivHand(g.GetPlayer(2), NewCard(CardDesignClover, 7, false))                                    // 7
	setYanivHand(g.GetPlayer(3), NewCard(CardDesignDiamond, 9, false))                                   // 9

	require.NoError(t, g.PlayerDeclareYaniv())
	assert.False(t, g.GetIsAsaf())
	assert.Equal(t, 0, g.GetCallerIdx())
	assert.Equal(t, 0, g.GetPlayer(0).GetScore())
	assert.Equal(t, 10, g.GetPlayer(1).GetScore())
	assert.Equal(t, 7, g.GetPlayer(2).GetScore())
	assert.Equal(t, 9, g.GetPlayer(3).GetScore())
	assert.Equal(t, YanivPhaseRoundEnd, g.GetPhase())
}

func TestYaniv_DeclareYanivAsaf(t *testing.T) {
	g := newTestYaniv()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(YanivPhaseDiscard)
	setYanivHand(g.GetPlayer(0), NewCard(CardDesignSpade, 5, false))   // caller = 5
	setYanivHand(g.GetPlayer(1), NewCard(CardDesignHeart, 3, false))   // 3 -> undercut
	setYanivHand(g.GetPlayer(2), NewCard(CardDesignClover, 8, false))  // 8
	setYanivHand(g.GetPlayer(3), NewCard(CardDesignDiamond, 6, false)) // 6

	require.NoError(t, g.PlayerDeclareYaniv())
	assert.True(t, g.GetIsAsaf())
	assert.Equal(t, 1, g.GetAsafWinnerIdx())
	assert.Equal(t, YanivAsafPenalty, g.GetPlayer(0).GetScore()) // +30
	assert.Equal(t, 0, g.GetPlayer(1).GetScore())                // undercutter scores 0
	assert.Equal(t, 8, g.GetPlayer(2).GetScore())
	assert.Equal(t, 6, g.GetPlayer(3).GetScore())
}

func TestYaniv_DeclareYanivTieIsAsaf(t *testing.T) {
	g := newTestYaniv()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(YanivPhaseDiscard)
	setYanivHand(g.GetPlayer(0), NewCard(CardDesignSpade, 5, false))   // 5
	setYanivHand(g.GetPlayer(1), NewCard(CardDesignHeart, 5, false))   // 5 -> tie counts as asaf
	setYanivHand(g.GetPlayer(2), NewCard(CardDesignClover, 9, false))  // 9
	setYanivHand(g.GetPlayer(3), NewCard(CardDesignDiamond, 9, false)) // 9
	require.NoError(t, g.PlayerDeclareYaniv())
	assert.True(t, g.GetIsAsaf())
	assert.Equal(t, 1, g.GetAsafWinnerIdx())
}

func TestYaniv_DeclareYanivTooHighErrors(t *testing.T) {
	g := newTestYaniv()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(YanivPhaseDiscard)
	setYanivHand(g.GetPlayer(0), NewCard(CardDesignSpade, 10, false))
	assert.Error(t, g.PlayerDeclareYaniv())
}

func TestYaniv_ResolveYanivEliminatesOverLimit(t *testing.T) {
	g := newTestYaniv()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(YanivPhaseDiscard)
	g.GetPlayer(1).SetScore(200)                                      // already at the limit
	setYanivHand(g.GetPlayer(0), NewCard(CardDesignSpade, 1, false))  // caller = 1
	setYanivHand(g.GetPlayer(1), NewCard(CardDesignHeart, 10, false)) // +10 -> 210 > 200 eliminated
	setYanivHand(g.GetPlayer(2), NewCard(CardDesignClover, 9, false))
	setYanivHand(g.GetPlayer(3), NewCard(CardDesignDiamond, 9, false))
	require.NoError(t, g.PlayerDeclareYaniv())
	assert.True(t, g.GetPlayer(1).IsEliminated())
}

func TestYaniv_GuardsWhenGameEnded(t *testing.T) {
	g := newTestYaniv()
	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlayerDiscard([]int{0}), ErrGameEnded)
	assert.ErrorIs(t, g.PlayerDeclareYaniv(), ErrGameEnded)
	assert.ErrorIs(t, g.PlayerDrawFromStock(), ErrGameEnded)
	assert.ErrorIs(t, g.PlayerDrawFromPickup(0), ErrGameEnded)
}

func TestYaniv_WrongPhaseGuards(t *testing.T) {
	g := newTestYaniv()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(YanivPhaseDraw)
	assert.ErrorIs(t, g.PlayerDiscard([]int{0}), ErrWrongPhase)
	assert.ErrorIs(t, g.PlayerDeclareYaniv(), ErrWrongPhase)
	g.SetPhase(YanivPhaseDiscard)
	assert.ErrorIs(t, g.PlayerDrawFromStock(), ErrWrongPhase)
	assert.ErrorIs(t, g.PlayerDrawFromPickup(0), ErrWrongPhase)
}

func TestYaniv_NotHumanTurn(t *testing.T) {
	g := newTestYaniv()
	g.SetCurrentPlayerIdx(1)
	g.SetPhase(YanivPhaseDiscard)
	assert.ErrorIs(t, g.PlayerDiscard([]int{0}), ErrNotHumanTurn)
	assert.ErrorIs(t, g.PlayerDeclareYaniv(), ErrNotHumanTurn)
	g.SetPhase(YanivPhaseDraw)
	assert.ErrorIs(t, g.PlayerDrawFromStock(), ErrNotHumanTurn)
}

func TestYaniv_DrawStockEmptyNoContest(t *testing.T) {
	g := newTestYaniv()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(YanivPhaseDraw)
	g.SetDrawPile([]*Card{})
	g.deadPile = nil
	g.pendingDiscard = []*Card{NewCard(CardDesignHeart, 9, false)}
	g.SetPickupCards([]*Card{NewCard(CardDesignSpade, 2, false)})
	require.NoError(t, g.PlayerDrawFromStock())
	assert.Contains(t, []YanivPhase{YanivPhaseRoundEnd, YanivPhaseGameEnd}, g.GetPhase())
	assert.Equal(t, -1, g.GetCallerIdx())
}

func TestYaniv_DrawStockReshufflesDeadPile(t *testing.T) {
	g := newTestYaniv()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(YanivPhaseDraw)
	g.SetDrawPile([]*Card{})
	g.deadPile = []*Card{NewCard(CardDesignClover, 4, false), NewCard(CardDesignDiamond, 5, false)}
	g.pendingDiscard = []*Card{NewCard(CardDesignHeart, 9, false)}
	g.SetPickupCards([]*Card{NewCard(CardDesignSpade, 2, false)})
	before := g.GetPlayer(0).GetCardsSize()
	require.NoError(t, g.PlayerDrawFromStock())
	assert.Equal(t, before+1, g.GetPlayer(0).GetCardsSize())
}

func TestYaniv_EliminationAndGameEnd(t *testing.T) {
	g := newTestYaniv()
	for i := 1; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetEliminated(true)
	}
	g.checkGameEnd()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerIdx())
}

func TestYaniv_HumanEliminationEndsGame(t *testing.T) {
	g := newTestYaniv()
	g.GetPlayer(0).SetEliminated(true)
	g.GetPlayer(1).SetScore(5)
	g.GetPlayer(2).SetScore(20)
	g.GetPlayer(3).SetScore(30)
	g.checkGameEnd()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetWinnerIdx()) // lowest score among survivors
}

func TestYaniv_NextRound(t *testing.T) {
	g := newTestYaniv()
	g.SetPhase(YanivPhaseRoundEnd)
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, YanivPhaseDiscard, g.GetPhase())

	// NextRound is a no-op outside RoundEnd phase.
	prev := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, prev, g.GetRoundNumber())
}

func TestYaniv_BestYanivDiscard(t *testing.T) {
	// A run of three spades (4+5+6=15) beats any single and the pair of 8s (16?).
	cards := []*Card{
		NewCard(CardDesignSpade, 4, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 6, false),
		NewCard(CardDesignHeart, 8, false),
		NewCard(CardDesignClover, 8, false),
	}
	idx := bestYanivDiscard(cards)
	// Pair of 8s removes 16 which is the max, so it should be chosen.
	assert.Len(t, idx, 2)

	// With only singles, the highest card is dumped.
	only := []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 10, false),
	}
	idx2 := bestYanivDiscard(only)
	require.Len(t, idx2, 1)
	assert.Equal(t, 10, only[idx2[0]].GetValue())
}

func TestYaniv_CpuPlaysToCompletion(t *testing.T) {
	for diff := YanivCpuDifficultyEasy; diff <= YanivCpuDifficultyHard; diff++ {
		g := NewDefaultYaniv()
		g.SetRand(rand.New(rand.NewSource(int64(diff) + 1)))
		cfg := DefaultYanivConfig()
		cfg.CpuDifficulty = diff
		cfg.ScoreLimit = 60 // shorten the game
		g.SetConfig(cfg)
		g.Reset()

		for step := 0; step < 20000 && !g.GetGameEndFlag(); step++ {
			switch {
			case g.GetPhase() == YanivPhaseRoundEnd:
				g.NextRound()
			case g.IsHumanTurn() && g.GetPhase() == YanivPhaseDiscard:
				if g.GetPlayer(0).HandTotal() <= YanivCallThreshold {
					_ = g.PlayerDeclareYaniv()
				} else {
					_ = g.PlayerDiscard(bestYanivDiscard(g.handCards(0)))
				}
			case g.IsHumanTurn() && g.GetPhase() == YanivPhaseDraw:
				_ = g.PlayerDrawFromStock()
			default:
				g.CpuPlay()
			}
		}
		assert.True(t, g.GetGameEndFlag(), "difficulty %d should finish", diff)
	}
}

func TestYaniv_CpuDiscardEmptyHandDoesNotPanic(t *testing.T) {
	g := newTestYaniv()
	g.SetCurrentPlayerIdx(1) // a CPU
	g.SetPhase(YanivPhaseDiscard)
	g.GetPlayer(1).Reset() // empty hand
	assert.NotPanics(t, func() { g.CpuPlay() })
}

func TestYaniv_JSONRoundTrip(t *testing.T) {
	g := newTestYaniv()
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored Yaniv
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetDrawPileCount(), restored.GetDrawPileCount())
	assert.Equal(t, len(g.GetPickupCards()), len(restored.GetPickupCards()))
}

func TestYaniv_UnmarshalRejectsOversize(t *testing.T) {
	huge := make([]int, yanivMaxSliceLen+1)
	payload, err := json.Marshal(yanivJSON{RoundScores: huge})
	require.NoError(t, err)
	var g Yaniv
	assert.Error(t, json.Unmarshal(payload, &g))
}

func TestYaniv_UnmarshalRejectsBadPlayerCount(t *testing.T) {
	for _, body := range []string{`{}`, `{"pl":[]}`} {
		var g Yaniv
		assert.Error(t, json.Unmarshal([]byte(body), &g), "body %q should be rejected", body)
	}
}

func TestYaniv_UnmarshalNilSliceDefaults(t *testing.T) {
	players := make([]*YanivPlayer, YanivPlayerCnt)
	for i := range players {
		players[i] = NewYanivPlayer(i == 0)
	}
	payload, err := json.Marshal(yanivJSON{Players: players})
	require.NoError(t, err)

	var g Yaniv
	require.NoError(t, json.Unmarshal(payload, &g))
	assert.NotNil(t, g.GetActionLog())
	assert.Equal(t, 0, g.GetDrawPileCount())
	assert.Empty(t, g.GetPickupCards())
	assert.Equal(t, YanivPlayerCnt, g.GetPlayerCnt())
}

func TestYaniv_ActionLogRecorded(t *testing.T) {
	g := newTestYaniv()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(YanivPhaseDiscard)
	setYanivHand(g.GetPlayer(0), NewCard(CardDesignSpade, 9, false), NewCard(CardDesignHeart, 3, false))
	g.SetDrawPile([]*Card{NewCard(CardDesignClover, 2, false)})
	require.NoError(t, g.PlayerDiscard([]int{0}))
	assert.NotEmpty(t, g.GetActionLog())
}

func TestYaniv_CardsStrHelper(t *testing.T) {
	assert.Equal(t, "-", cardsStr(nil))
	out := cardsStr([]*Card{NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 2, false)})
	assert.NotEmpty(t, out)
}
