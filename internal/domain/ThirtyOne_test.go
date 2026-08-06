package domain

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestThirtyOne builds a 4-player game with a deterministic RNG and the
// default config, then resets it so a round is in progress.
func newTestThirtyOne() *ThirtyOne {
	g := NewDefaultThirtyOne()
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()
	return g
}

// setThirtyOneHand replaces a player's hand with the given cards.
func setThirtyOneHand(p *ThirtyOnePlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestNewThirtyOne_Defaults(t *testing.T) {
	g := NewDefaultThirtyOne()
	assert.Equal(t, ThirtyOnePlayerCnt, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman())
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Nil(t, g.GetPlayer(99))
}

func TestThirtyOne_ResetDealsHands(t *testing.T) {
	g := newTestThirtyOne()
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, ThirtyOnePhaseDraw, g.GetPhase())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		// Blitz on deal can short-circuit a round, but with seed 1 no player
		// is dealt 31, so everyone keeps 3 cards.
		if g.GetPhase() == ThirtyOnePhaseDraw {
			assert.Equal(t, ThirtyOneHandSize, g.GetPlayer(i).GetCardsSize())
			assert.Equal(t, 3, g.GetPlayer(i).GetLives())
		}
	}
	assert.NotNil(t, g.GetDiscardTop())
	assert.Positive(t, g.GetDrawPileCount())
}

func TestThirtyOnePlayer_SuitScoring(t *testing.T) {
	p := NewThirtyOnePlayer(true)
	setThirtyOneHand(p,
		NewCard(CardDesignSpade, 1, false),  // 11
		NewCard(CardDesignSpade, 13, false), // 10
		NewCard(CardDesignSpade, 12, false), // 10
	)
	assert.Equal(t, 31, p.BestSuitScore())
	assert.Equal(t, CardDesignSpade, p.BestSuit())

	setThirtyOneHand(p,
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignClover, 2, false),
	)
	assert.Equal(t, 9, p.BestSuitScore())
	assert.Equal(t, CardDesignDiamond, p.BestSuit())
}

func TestThirtyOne_HumanDrawStockThenDiscard(t *testing.T) {
	g := newTestThirtyOne()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThirtyOnePhaseDraw)
	before := g.GetPlayer(0).GetCardsSize()
	require.NoError(t, g.PlayerDrawFromStock())
	assert.Equal(t, ThirtyOnePhaseDiscard, g.GetPhase())
	assert.Equal(t, before+1, g.GetPlayer(0).GetCardsSize())

	require.NoError(t, g.PlayerDiscard(0))
	assert.Equal(t, before, g.GetPlayer(0).GetCardsSize())
}

func TestThirtyOne_HumanDrawFromDiscard(t *testing.T) {
	g := newTestThirtyOne()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThirtyOnePhaseDraw)
	top := g.GetDiscardTop()
	require.NotNil(t, top)
	require.NoError(t, g.PlayerDrawFromDiscard())
	assert.Equal(t, ThirtyOnePhaseDiscard, g.GetPhase())
}

func TestThirtyOne_DrawFromEmptyDiscardErrors(t *testing.T) {
	g := newTestThirtyOne()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThirtyOnePhaseDraw)
	g.SetDiscardPile([]*Card{})
	assert.Error(t, g.PlayerDrawFromDiscard())
}

func TestThirtyOne_DrawStockEmptyEndsRound(t *testing.T) {
	g := newTestThirtyOne()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThirtyOnePhaseDraw)
	g.SetDrawPile([]*Card{})
	require.NoError(t, g.PlayerDrawFromStock())
	assert.Contains(t, []ThirtyOnePhase{ThirtyOnePhaseRoundEnd, ThirtyOnePhaseGameEnd}, g.GetPhase())
}

func TestThirtyOne_DiscardWrongPhaseAndIndex(t *testing.T) {
	g := newTestThirtyOne()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThirtyOnePhaseDraw)
	assert.ErrorIs(t, g.PlayerDiscard(0), ErrWrongPhase)

	g.SetPhase(ThirtyOnePhaseDiscard)
	assert.Error(t, g.PlayerDiscard(99))
}

func TestThirtyOne_DiscardReaching31WinsRound(t *testing.T) {
	g := newTestThirtyOne()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThirtyOnePhaseDiscard)
	// Hand of 4: A♠ K♠ Q♠ (=31) + a junk card to discard.
	setThirtyOneHand(g.GetPlayer(0),
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignHeart, 2, false),
	)
	require.NoError(t, g.PlayerDiscard(3))
	assert.Equal(t, 0, g.GetThirtyOneIdx())
	assert.Equal(t, 0, g.GetRoundWinnerIdx())
	// Every other active player lost a life.
	assert.Len(t, g.GetRoundLosers(), 3)
}

func TestThirtyOne_KnockFlowEndsRoundAfterFullLoop(t *testing.T) {
	g := newTestThirtyOne()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThirtyOnePhaseDraw)
	require.NoError(t, g.PlayerKnock())
	assert.Equal(t, 0, g.GetKnockerIdx())
	// Knocking moves the turn forward; the human is no longer to act.
	assert.NotEqual(t, 0, g.GetCurrentPlayerIdx())

	// Re-knocking is rejected.
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThirtyOnePhaseDraw)
	assert.Error(t, g.PlayerKnock())
}

func TestThirtyOne_KnockWrongPhase(t *testing.T) {
	g := newTestThirtyOne()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThirtyOnePhaseDiscard)
	assert.ErrorIs(t, g.PlayerKnock(), ErrWrongPhase)
}

func TestThirtyOne_EndRoundLowestLosesLife(t *testing.T) {
	g := newTestThirtyOne()
	// Force deterministic scores: p0 lowest.
	setThirtyOneHand(g.GetPlayer(0), NewCard(CardDesignHeart, 2, false))   // 2
	setThirtyOneHand(g.GetPlayer(1), NewCard(CardDesignSpade, 1, false))   // 11
	setThirtyOneHand(g.GetPlayer(2), NewCard(CardDesignClover, 10, false)) // 10
	setThirtyOneHand(g.GetPlayer(3), NewCard(CardDesignDiamond, 9, false)) // 9
	g.SetKnockerIdx(1)
	g.SetCurrentPlayerIdx(1)
	g.SetPhase(ThirtyOnePhaseDraw)
	// Knocker is current; advancing around the table from the knocker ends it.
	g.advanceTurnForTest()
	// p0 should have lost a life.
	assert.Equal(t, 2, g.GetPlayer(0).GetLives())
	assert.Equal(t, []int{0}, g.GetRoundLosers())
	assert.Equal(t, 1, g.GetRoundWinnerIdx())
}

// advanceTurnForTest exercises the unexported round-advance path through a knock
// loop by repeatedly advancing until the round resolves.
func (g *ThirtyOne) advanceTurnForTest() {
	for i := 0; i < ThirtyOnePlayerCnt+1; i++ {
		if g.GetPhase() == ThirtyOnePhaseRoundEnd || g.GetPhase() == ThirtyOnePhaseGameEnd {
			return
		}
		g.advanceTurn()
	}
}

func TestThirtyOne_NextRound(t *testing.T) {
	g := newTestThirtyOne()
	g.GetPlayer(0).LoseLife() // make a loser scenario survive
	// Drive a round end via stock-out.
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThirtyOnePhaseDraw)
	g.SetDrawPile([]*Card{})
	require.NoError(t, g.PlayerDrawFromStock())
	if g.GetPhase() == ThirtyOnePhaseRoundEnd {
		g.NextRound()
		assert.Equal(t, 2, g.GetRoundNumber())
		assert.Equal(t, ThirtyOnePhaseDraw, g.GetPhase())
	}
	// NextRound is a no-op outside RoundEnd phase.
	g.SetPhase(ThirtyOnePhaseDraw)
	prev := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, prev, g.GetRoundNumber())
}

func TestThirtyOne_EliminationAndGameEnd(t *testing.T) {
	g := newTestThirtyOne()
	// Drop everyone but the human to eliminated.
	for i := 1; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetLives(-1)
	}
	g.checkGameEnd()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerIdx())
}

func TestThirtyOne_HumanEliminationEndsGame(t *testing.T) {
	g := newTestThirtyOne()
	g.GetPlayer(0).SetLives(-1)
	g.checkGameEnd()
	assert.True(t, g.GetGameEndFlag())
	// Winner is the player with the most lives (a CPU).
	assert.NotEqual(t, 0, g.GetWinnerIdx())
}

func TestThirtyOne_GuardsWhenGameEnded(t *testing.T) {
	g := newTestThirtyOne()
	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlayerDrawFromStock(), ErrGameEnded)
	assert.ErrorIs(t, g.PlayerDrawFromDiscard(), ErrGameEnded)
	assert.ErrorIs(t, g.PlayerDiscard(0), ErrGameEnded)
	assert.ErrorIs(t, g.PlayerKnock(), ErrGameEnded)
}

func TestThirtyOne_NotHumanTurn(t *testing.T) {
	g := newTestThirtyOne()
	g.SetCurrentPlayerIdx(1)
	g.SetPhase(ThirtyOnePhaseDraw)
	assert.ErrorIs(t, g.PlayerDrawFromStock(), ErrNotHumanTurn)
	g.SetPhase(ThirtyOnePhaseDiscard)
	assert.ErrorIs(t, g.PlayerDiscard(0), ErrNotHumanTurn)
}

func TestThirtyOne_CpuPlaysToCompletion(t *testing.T) {
	for diff := ThirtyOneCpuDifficultyEasy; diff <= ThirtyOneCpuDifficultyHard; diff++ {
		g := NewDefaultThirtyOne()
		g.SetRand(rand.New(rand.NewSource(int64(diff) + 1)))
		cfg := DefaultThirtyOneConfig()
		cfg.CpuDifficulty = diff
		cfg.InitialLives = 1
		g.SetConfig(cfg)
		g.Reset()

		// Run many CPU turns; whenever it is the human's turn, just draw+discard.
		for step := 0; step < 5000 && !g.GetGameEndFlag(); step++ {
			switch {
			case g.GetPhase() == ThirtyOnePhaseRoundEnd:
				g.NextRound()
			case g.IsHumanTurn() && g.GetPhase() == ThirtyOnePhaseDraw:
				_ = g.PlayerDrawFromStock()
			case g.IsHumanTurn() && g.GetPhase() == ThirtyOnePhaseDiscard:
				_ = g.PlayerDiscard(0)
			default:
				g.CpuPlay()
			}
		}
		assert.True(t, g.GetGameEndFlag(), "difficulty %d should finish", diff)
	}
}

func TestThirtyOne_CpuDiscardEmptyHandDoesNotPanic(t *testing.T) {
	g := newTestThirtyOne()
	g.SetCurrentPlayerIdx(1) // a CPU
	g.SetPhase(ThirtyOnePhaseDiscard)
	g.GetPlayer(1).Reset() // empty hand
	assert.NotPanics(t, func() { g.CpuPlay() })
}

func TestThirtyOne_BlitzOnDeal(t *testing.T) {
	// Hand-craft a draw pile so the human is dealt A♠ K♠ Q♠ = 31.
	g := NewDefaultThirtyOne()
	g.SetRand(rand.New(rand.NewSource(1)))
	// Reset normally, then re-run the blitz check with a forced 31 hand.
	g.Reset()
	g.SetPhase(ThirtyOnePhaseDraw)
	setThirtyOneHand(g.GetPlayer(0),
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 12, false),
	)
	g.checkBlitzOnDeal()
	assert.Equal(t, 0, g.GetThirtyOneIdx())
}

func TestThirtyOne_JSONRoundTrip(t *testing.T) {
	g := newTestThirtyOne()
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored ThirtyOne
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetDrawPileCount(), restored.GetDrawPileCount())
}

func TestThirtyOne_UnmarshalRejectsOversize(t *testing.T) {
	huge := make([]int, thirtyOneMaxSliceLen+1)
	payload, err := json.Marshal(thirtyOneJSON{RoundLosers: huge})
	require.NoError(t, err)
	var g ThirtyOne
	assert.Error(t, json.Unmarshal(payload, &g))
}

func TestThirtyOne_UnmarshalRejectsBadPlayerCount(t *testing.T) {
	// Empty / wrong-sized players slices are rejected to prevent downstream panics.
	for _, body := range []string{`{}`, `{"pl":[]}`} {
		var g ThirtyOne
		assert.Error(t, json.Unmarshal([]byte(body), &g), "body %q should be rejected", body)
	}
}

func TestThirtyOne_UnmarshalNilSliceDefaults(t *testing.T) {
	// A valid 4-player payload with nil collection fields gets safe defaults.
	players := make([]*ThirtyOnePlayer, ThirtyOnePlayerCnt)
	for i := range players {
		players[i] = NewThirtyOnePlayer(i == 0)
	}
	payload, err := json.Marshal(thirtyOneJSON{Players: players})
	require.NoError(t, err)

	var g ThirtyOne
	require.NoError(t, json.Unmarshal(payload, &g))
	assert.NotNil(t, g.GetActionLog())
	assert.Equal(t, 0, g.GetDrawPileCount())
	assert.Empty(t, g.GetRoundLosers())
	assert.Equal(t, ThirtyOnePlayerCnt, g.GetPlayerCnt())
}

func TestThirtyOne_BestDropHelpers(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignSpade, 1, false),  // 11
		NewCard(CardDesignSpade, 13, false), // 10
		NewCard(CardDesignSpade, 5, false),  // 5
		NewCard(CardDesignHeart, 2, false),  // 2
	}
	// Dropping the heart keeps 26 in spades.
	assert.Equal(t, 26, bestScoreAfterDrop(cards))
	assert.Equal(t, 3, bestDropIndex(cards))
}

func TestThirtyOne_ActionLogRecorded(t *testing.T) {
	g := newTestThirtyOne()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThirtyOnePhaseDraw)
	require.NoError(t, g.PlayerDrawFromStock())
	assert.NotEmpty(t, g.GetActionLog())
}

// **CPU と同じ材料で判断すること (#4806)。**別の計算を書くと、CPU には有利と
// 見える手を人間には勧めない、という食い違いが出る。
func TestThirtyOne_GetHint(t *testing.T) {
	g := NewDefaultThirtyOne()
	g.Reset()

	// CPU の手番では nil。
	g.SetCurrentPlayerIdx(1)
	assert.Nil(t, g.GetHint())

	// ディスカードフェーズは捨てる札を指す。
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThirtyOnePhaseDiscard)
	h := g.GetHint()
	assert.NotNil(t, h)
	assert.Equal(t, "discard", h.Action)
	assert.GreaterOrEqual(t, h.CardIndex, 0)
	assert.Less(t, h.CardIndex, g.GetPlayer(0).GetCardsSize())

	// ドローフェーズは 3 つの行動のいずれか。
	g.SetPhase(ThirtyOnePhaseDraw)
	h2 := g.GetHint()
	assert.NotNil(t, h2)
	assert.Contains(t, []string{"draw_stock", "draw_discard", "knock"}, h2.Action)
	assert.Equal(t, -1, h2.CardIndex)

	// 終了フェーズでは行動を返さない。
	g.SetPhase(ThirtyOnePhaseGameEnd)
	assert.Nil(t, g.GetHint())
}
