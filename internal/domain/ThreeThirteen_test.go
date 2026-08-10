//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ttCard builds a *Card with the given design and value.
func ttCard(d, v int) *Card { return NewCard(d, v, false) }

// ttSetHand replaces a player's hand with the given cards.
func ttSetHand(p *ThreeThirteenPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func newTestThreeThirteen(playerCount int) *ThreeThirteen {
	cfg := DefaultThreeThirteenConfig()
	cfg.PlayerCount = playerCount
	return NewThreeThirteen(NewTrumpCardsWithDecks(2, 0), buildThreeThirteenPlayers(playerCount), cfg)
}

func TestThreeThirteen_DeckIs104(t *testing.T) {
	tc := NewTrumpCardsWithDecks(2, 0)
	assert.Equal(t, 104, tc.GetTotalCount())
}

func TestThreeThirteen_WildRankAndDealCountPerRound(t *testing.T) {
	for round := ThreeThirteenMinRound; round <= ThreeThirteenMaxRound; round++ {
		assert.Equal(t, round+2, ThreeThirteenWildRankFor(round))
		assert.Equal(t, round+2, ThreeThirteenDealCountFor(round))
	}
	assert.Equal(t, 3, ThreeThirteenWildRankFor(1))   // round 1 → rank 3 wild
	assert.Equal(t, 13, ThreeThirteenWildRankFor(11)) // round 11 → rank 13 (K) wild
}

func TestThreeThirteen_ResetDealsCorrectCount(t *testing.T) {
	g := newTestThreeThirteen(4)
	g.Reset()
	assert.Equal(t, 1, g.GetRound())
	assert.Equal(t, ThreeThirteenPhaseDraw, g.GetPhase())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, 3, g.GetPlayer(i).GetCardsSize(), "round 1 deals 3 cards")
	}
	assert.NotNil(t, g.GetDiscardTop())
	// 104 - (4*3) - 1 discard = 91
	assert.Equal(t, 91, g.GetDrawPileCount())
}

func TestThreeThirteen_IsWild(t *testing.T) {
	assert.True(t, threeThirteenIsWild(ttCard(CardDesignSpade, 5), 5))
	assert.False(t, threeThirteenIsWild(ttCard(CardDesignSpade, 6), 5))
	assert.False(t, threeThirteenIsWild(nil, 5))
}

func TestThreeThirteen_IsSet(t *testing.T) {
	wild := 3
	// natural set of 3
	assert.True(t, threeThirteenIsValidMeld([]*Card{ttCard(CardDesignSpade, 7), ttCard(CardDesignHeart, 7), ttCard(CardDesignDiamond, 7)}, wild))
	// set with a wild substituting
	assert.True(t, threeThirteenIsValidMeld([]*Card{ttCard(CardDesignSpade, 7), ttCard(CardDesignHeart, 7), ttCard(CardDesignSpade, 3)}, wild))
	// not a set
	assert.False(t, threeThirteenIsValidMeld([]*Card{ttCard(CardDesignSpade, 7), ttCard(CardDesignHeart, 8), ttCard(CardDesignDiamond, 9)}, wild))
	// all-wild is not a set
	assert.False(t, threeThirteenIsSet([]*Card{ttCard(CardDesignSpade, 3), ttCard(CardDesignHeart, 3), ttCard(CardDesignDiamond, 3)}, 3))
}

func TestThreeThirteen_IsRun(t *testing.T) {
	wild := 13
	// natural run
	assert.True(t, threeThirteenIsValidMeld([]*Card{ttCard(CardDesignSpade, 4), ttCard(CardDesignSpade, 5), ttCard(CardDesignSpade, 6)}, wild))
	// run with a gap filled by a wild (K is wild this round)
	assert.True(t, threeThirteenIsValidMeld([]*Card{ttCard(CardDesignSpade, 4), ttCard(CardDesignSpade, 13), ttCard(CardDesignSpade, 6)}, wild))
	// run with ace-low
	assert.True(t, threeThirteenIsValidMeld([]*Card{ttCard(CardDesignHeart, 1), ttCard(CardDesignHeart, 2), ttCard(CardDesignHeart, 3)}, 13))
	// mixed suit is not a run
	assert.False(t, threeThirteenIsRun([]*Card{ttCard(CardDesignSpade, 4), ttCard(CardDesignHeart, 5), ttCard(CardDesignSpade, 6)}, wild))
	// duplicate value cannot form a run
	assert.False(t, threeThirteenIsRun([]*Card{ttCard(CardDesignSpade, 4), ttCard(CardDesignSpade, 4), ttCard(CardDesignSpade, 5)}, wild))
}

func TestThreeThirteen_BestMeldsFullMeld(t *testing.T) {
	wild := 3
	// 6 cards: a set of 7s and a run of spades → deadwood 0
	cards := []*Card{
		ttCard(CardDesignSpade, 7), ttCard(CardDesignHeart, 7), ttCard(CardDesignDiamond, 7),
		ttCard(CardDesignSpade, 9), ttCard(CardDesignSpade, 10), ttCard(CardDesignSpade, 11),
	}
	melds, deadwood := threeThirteenBestMelds(cards, wild)
	assert.Equal(t, 0, threeThirteenDeadwoodValue(deadwood, wild))
	assert.Len(t, melds, 2)
}

func TestThreeThirteen_DeadwoodValue(t *testing.T) {
	wild := 3
	// A=1, 10=10, K=10, wild(3)=20
	dead := []*Card{ttCard(CardDesignSpade, 1), ttCard(CardDesignHeart, 10), ttCard(CardDesignDiamond, 13), ttCard(CardDesignClover, 3)}
	assert.Equal(t, 1+10+10+ThreeThirteenWildDeadwoodValue, threeThirteenDeadwoodValue(dead, wild))
}

func TestThreeThirteen_DrawStockThenDiscard(t *testing.T) {
	g := newTestThreeThirteen(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThreeThirteenPhaseDraw)
	g.SetDrawPile([]*Card{ttCard(CardDesignSpade, 8)})
	before := g.GetPlayer(0).GetCardsSize()

	require.NoError(t, g.PlayerDrawFromStock())
	assert.Equal(t, before+1, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, ThreeThirteenPhaseDiscard, g.GetPhase())

	require.NoError(t, g.PlayerDiscard(0))
	assert.Equal(t, before, g.GetPlayer(0).GetCardsSize())
	// turn advanced to player 1
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
}

func TestThreeThirteen_DrawFromDiscard(t *testing.T) {
	g := newTestThreeThirteen(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThreeThirteenPhaseDraw)
	g.SetDiscardPile([]*Card{ttCard(CardDesignHeart, 4)})
	require.NoError(t, g.PlayerDrawFromDiscard())
	assert.Equal(t, ThreeThirteenPhaseDiscard, g.GetPhase())
	assert.Nil(t, g.GetDiscardTop())
}

func TestThreeThirteen_WrongPhaseAndTurnGuards(t *testing.T) {
	g := newTestThreeThirteen(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThreeThirteenPhaseDiscard)
	assert.ErrorIs(t, g.PlayerDrawFromStock(), ErrWrongPhase)

	g.SetPhase(ThreeThirteenPhaseDraw)
	g.SetCurrentPlayerIdx(1) // CPU
	assert.ErrorIs(t, g.PlayerDrawFromStock(), ErrNotHumanTurn)

	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlayerDrawFromStock(), ErrGameEnded)
	assert.ErrorIs(t, g.PlayerDiscard(0), ErrGameEnded)
}

func TestThreeThirteen_DiscardOutOfRange(t *testing.T) {
	g := newTestThreeThirteen(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThreeThirteenPhaseDiscard)
	err := g.PlayerDiscard(999)
	assert.ErrorIs(t, err, ErrInvalidCard)
}

func TestThreeThirteen_KnockRejectedWithDeadwood(t *testing.T) {
	g := newTestThreeThirteen(2)
	g.Reset()
	g.SetRound(1) // wild = 3
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThreeThirteenPhaseDiscard)
	// hand of 4: after discarding one, 3 cards remain that do NOT meld
	ttSetHand(g.GetPlayer(0), ttCard(CardDesignSpade, 5), ttCard(CardDesignHeart, 8), ttCard(CardDesignDiamond, 11), ttCard(CardDesignClover, 2))
	err := g.PlayerKnock(0)
	assert.ErrorIs(t, err, ErrInvalidPlay)
}

func TestThreeThirteen_KnockAcceptedWhenFullyMelded(t *testing.T) {
	g := newTestThreeThirteen(2)
	g.Reset()
	g.SetRound(1) // wild = 3, deal count 3 (round 1) but we set hand manually to 4
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThreeThirteenPhaseDiscard)
	g.knockerIdx = -1
	// 4 cards: a set of 8s (3) + 1 junk to discard
	ttSetHand(g.GetPlayer(0),
		ttCard(CardDesignSpade, 8), ttCard(CardDesignHeart, 8), ttCard(CardDesignDiamond, 8),
		ttCard(CardDesignClover, 11))
	require.NoError(t, g.PlayerKnock(3))
	assert.Equal(t, 0, g.GetKnockerIdx())
}

func TestThreeThirteen_RoundScoringAfterKnock(t *testing.T) {
	g := newTestThreeThirteen(2)
	g.Reset()
	g.SetRound(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThreeThirteenPhaseDiscard)
	g.knockerIdx = -1
	// knocker fully melds 8-8-8 after discarding
	ttSetHand(g.GetPlayer(0),
		ttCard(CardDesignSpade, 8), ttCard(CardDesignHeart, 8), ttCard(CardDesignDiamond, 8),
		ttCard(CardDesignClover, 11))
	// opponent has deadwood
	ttSetHand(g.GetPlayer(1), ttCard(CardDesignSpade, 5), ttCard(CardDesignHeart, 9), ttCard(CardDesignDiamond, 12))
	g.SetDrawPile([]*Card{ttCard(CardDesignSpade, 2)})

	require.NoError(t, g.PlayerKnock(3))
	// player 1 (CPU) takes its final turn via CpuPlay loop simulated here:
	for !g.GetGameEndFlag() && g.GetPhase() != ThreeThirteenPhaseRoundEnd {
		if g.IsHumanTurn() {
			break
		}
		g.CpuPlay()
	}
	assert.Equal(t, ThreeThirteenPhaseRoundEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetPlayer(0).GetCumulativeScore(), "knocker melds fully → 0")
	assert.Greater(t, g.GetPlayer(1).GetCumulativeScore(), 0)
}

func TestThreeThirteen_NextRoundProgressesWildAndDeal(t *testing.T) {
	g := newTestThreeThirteen(2)
	g.Reset()
	assert.Equal(t, 1, g.GetRound())
	g.SetPhase(ThreeThirteenPhaseRoundEnd)
	g.NextRound()
	assert.Equal(t, 2, g.GetRound())
	assert.Equal(t, 4, g.WildRank())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, 4, g.GetPlayer(i).GetCardsSize())
	}
}

func TestThreeThirteen_NextRoundIgnoredOutsideRoundEnd(t *testing.T) {
	g := newTestThreeThirteen(2)
	g.Reset()
	g.SetPhase(ThreeThirteenPhaseDraw)
	g.NextRound()
	assert.Equal(t, 1, g.GetRound(), "NextRound is a no-op outside RoundEnd")
}

func TestThreeThirteen_GameEndsAfterRound11LowestWins(t *testing.T) {
	g := newTestThreeThirteen(2)
	g.Reset()
	g.SetRound(ThreeThirteenMaxRound)
	g.SetPhase(ThreeThirteenPhaseRoundEnd)
	g.GetPlayer(0).SetCumulativeScore(50)
	g.GetPlayer(1).SetCumulativeScore(10)
	g.NextRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, ThreeThirteenPhaseGameEnd, g.GetPhase())
	assert.Equal(t, 1, g.GetWinnerIdx(), "lowest cumulative score wins")
}

func TestThreeThirteen_StockOutRecycleAndEnd(t *testing.T) {
	g := newTestThreeThirteen(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThreeThirteenPhaseDraw)
	// empty stock + single discard → cannot recycle → round ends
	g.SetDrawPile(nil)
	g.SetDiscardPile([]*Card{ttCard(CardDesignSpade, 4)})
	require.NoError(t, g.PlayerDrawFromStock())
	assert.Equal(t, ThreeThirteenPhaseRoundEnd, g.GetPhase())
}

func TestThreeThirteen_DrawStockRecyclesDiscard(t *testing.T) {
	g := newTestThreeThirteen(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ThreeThirteenPhaseDraw)
	g.SetDrawPile(nil)
	g.SetDiscardPile([]*Card{ttCard(CardDesignSpade, 4), ttCard(CardDesignHeart, 5), ttCard(CardDesignDiamond, 6)})
	require.NoError(t, g.PlayerDrawFromStock())
	assert.Equal(t, ThreeThirteenPhaseDiscard, g.GetPhase())
}

func TestThreeThirteen_FullCpuGame_MultiConfig(t *testing.T) {
	difficulties := []ThreeThirteenCpuDifficulty{
		ThreeThirteenCpuDifficultyEasy,
		ThreeThirteenCpuDifficultyNormal,
		ThreeThirteenCpuDifficultyHard,
	}
	for _, diff := range difficulties {
		for pc := ThreeThirteenMinPlayers; pc <= ThreeThirteenMaxPlayers; pc++ {
			cfg := ThreeThirteenConfig{CpuDifficulty: diff, PlayerCount: pc}
			// all CPU players (override the human seat)
			players := make([]*ThreeThirteenPlayer, pc)
			for i := range players {
				players[i] = NewThreeThirteenPlayer(false)
			}
			g := NewThreeThirteen(NewTrumpCardsWithDecks(2, 0), players, cfg)
			g.Reset()

			guard := 0
			for !g.GetGameEndFlag() {
				guard++
				require.Less(t, guard, 200000, "game must terminate (diff=%d pc=%d)", diff, pc)
				phase := g.GetPhase()
				if phase == ThreeThirteenPhaseRoundEnd {
					g.NextRound()
					continue
				}
				if phase == ThreeThirteenPhaseGameEnd {
					break
				}
				g.CpuPlay()
			}
			assert.True(t, g.GetGameEndFlag())
			assert.Equal(t, ThreeThirteenMaxRound, g.GetRound())
			w := g.GetWinnerIdx()
			assert.GreaterOrEqual(t, w, 0)
			assert.Less(t, w, pc)
			// winner has the lowest cumulative score
			for i := 0; i < pc; i++ {
				assert.LessOrEqual(t, g.GetPlayer(w).GetCumulativeScore(), g.GetPlayer(i).GetCumulativeScore())
			}
		}
	}
}

func TestThreeThirteen_JSONRoundTrip(t *testing.T) {
	g := NewDefaultThreeThirteen()
	g.Reset()
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored ThreeThirteen
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetRound(), restored.GetRound())
	assert.Equal(t, g.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, g.GetPhase(), restored.GetPhase())
}

func TestThreeThirteen_UnmarshalRejectsInvalid(t *testing.T) {
	base := NewDefaultThreeThirteen()
	base.Reset()
	good, err := json.Marshal(base)
	require.NoError(t, err)

	t.Run("not json", func(t *testing.T) {
		var g ThreeThirteen
		assert.Error(t, g.UnmarshalJSON([]byte("nope")))
	})

	t.Run("bad round", func(t *testing.T) {
		var raw map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(good, &raw))
		raw["rd"] = json.RawMessage("99")
		b, _ := json.Marshal(raw)
		var g ThreeThirteen
		assert.Error(t, g.UnmarshalJSON(b))
	})

	t.Run("bad current player idx", func(t *testing.T) {
		var raw map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(good, &raw))
		raw["ci"] = json.RawMessage("99")
		b, _ := json.Marshal(raw)
		var g ThreeThirteen
		assert.Error(t, g.UnmarshalJSON(b))
	})

	t.Run("bad phase", func(t *testing.T) {
		var raw map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(good, &raw))
		raw["ps"] = json.RawMessage("99")
		b, _ := json.Marshal(raw)
		var g ThreeThirteen
		assert.Error(t, g.UnmarshalJSON(b))
	})

	t.Run("bad knocker idx", func(t *testing.T) {
		var raw map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(good, &raw))
		raw["kn"] = json.RawMessage("-5")
		b, _ := json.Marshal(raw)
		var g ThreeThirteen
		assert.Error(t, g.UnmarshalJSON(b))
	})
}

func TestThreeThirteenConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultThreeThirteenConfig().Validate())
	assert.Error(t, ThreeThirteenConfig{CpuDifficulty: -1, PlayerCount: 4}.Validate())
	assert.Error(t, ThreeThirteenConfig{CpuDifficulty: ThreeThirteenCpuDifficultyNormal, PlayerCount: 99}.Validate())
}

func TestThreeThirteenPlayer_JSONRoundTrip(t *testing.T) {
	p := NewThreeThirteenPlayer(true)
	p.AddCard(ttCard(CardDesignSpade, 5))
	p.SetCumulativeScore(13)
	data, err := json.Marshal(p)
	require.NoError(t, err)
	var restored ThreeThirteenPlayer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.True(t, restored.GetIsHuman())
	assert.Equal(t, 13, restored.GetCumulativeScore())
}

func TestThreeThirteen_GetPlayerDeadwoodValue(t *testing.T) {
	g := newTestThreeThirteen(2)
	g.Reset()
	g.SetRound(1)
	ttSetHand(g.GetPlayer(0), ttCard(CardDesignSpade, 8), ttCard(CardDesignHeart, 8), ttCard(CardDesignDiamond, 8))
	assert.Equal(t, 0, g.GetPlayerDeadwoodValue(0))
	assert.Equal(t, 0, g.GetPlayerDeadwoodValue(99), "out of range → 0")
}

// **捨てる前の予測。**Web の bestThreeThirteenDeadwoodValue と同じ値になること。
// ワイルドを残すと 20 点なので、ワイルドを捨てる手が有利になる場面がある (#4840)。
func TestThreeThirteen_GetDeadwoodAfterDiscard(t *testing.T) {
	g := NewDefaultThreeThirteen()
	g.Reset()
	p := g.GetPlayer(0)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	// ラウンド 1 のワイルドは 3。メルドにならない手札を作る。
	p.AddCard(NewCard(CardDesignSpade, 9, false))
	p.AddCard(NewCard(CardDesignHeart, 5, false))
	p.AddCard(NewCard(CardDesignClover, 3, false)) // ワイルド (残すと 20 点)

	assert.Equal(t, 3, g.WildRank())
	// 全部残すと 9 + 5 + 20 = 34。
	assert.Equal(t, 34, g.GetPlayerDeadwoodValue(0))
	// 9 を捨てると 5 + 20 = 25。
	assert.Equal(t, 25, g.GetDeadwoodAfterDiscard(0, 0))
	// ワイルドを捨てると 9 + 5 = 14 — いちばん減る。
	assert.Equal(t, 14, g.GetDeadwoodAfterDiscard(0, 2))
	// 範囲外は -1。
	assert.Equal(t, -1, g.GetDeadwoodAfterDiscard(0, 99))
	assert.Equal(t, -1, g.GetDeadwoodAfterDiscard(0, -1))
	assert.Equal(t, -1, g.GetDeadwoodAfterDiscard(99, 0))
}
