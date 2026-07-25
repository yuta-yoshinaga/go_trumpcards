//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newInternalCallBreak() *CallBreak {
	players := []*CallBreakPlayer{
		NewCallBreakPlayer(true),
		NewCallBreakPlayer(false),
		NewCallBreakPlayer(false),
		NewCallBreakPlayer(false),
	}
	return NewCallBreak(NewTrumpCards(0), players, DefaultCallBreakConfig())
}

func TestCallBreak_findHumanIdx(t *testing.T) {
	cb := newInternalCallBreak()
	assert.Equal(t, 0, cb.findHumanIdx())

	allCpu := []*CallBreakPlayer{
		NewCallBreakPlayer(false),
		NewCallBreakPlayer(false),
		NewCallBreakPlayer(false),
		NewCallBreakPlayer(false),
	}
	cb2 := NewCallBreak(NewTrumpCards(0), allCpu, DefaultCallBreakConfig())
	assert.Equal(t, -1, cb2.findHumanIdx())
}

func TestCallBreak_playerHasSuit(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	assert.True(t, cb.playerHasSuit(0, CardDesignSpade))
	assert.False(t, cb.playerHasSuit(0, CardDesignHeart))
}

func TestCallBreak_playerHasNonSpade(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	assert.False(t, cb.playerHasNonSpade(0))
	cb.players[0].AddCard(NewCard(CardDesignHeart, 3, false))
	assert.True(t, cb.playerHasNonSpade(0))
}

func TestCallBreak_trickWinner_EmptyTrick(t *testing.T) {
	cb := newInternalCallBreak()
	cb.currentTrick = nil
	assert.Equal(t, 0, cb.trickWinner())
}

func TestCallBreak_trickWinner_LeadSuitWins(t *testing.T) {
	cb := newInternalCallBreak()
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 5, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 11, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 4, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignClover, 13, false)},
	}
	assert.Equal(t, 1, cb.trickWinner())
}

func TestCallBreak_trickWinner_SpadeBeatsLeadSuit(t *testing.T) {
	cb := newInternalCallBreak()
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 13, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 2, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 12, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 11, false)},
	}
	assert.Equal(t, 1, cb.trickWinner())
}

func TestCallBreak_trickWinner_HigherSpadeWins(t *testing.T) {
	cb := newInternalCallBreak()
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 13, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 2, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 7, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 5, false)},
	}
	assert.Equal(t, 2, cb.trickWinner())
}

func TestCallBreak_trickWinner_AllSpades(t *testing.T) {
	cb := newInternalCallBreak()
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 3, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 11, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 7, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)},
	}
	assert.Equal(t, 3, cb.trickWinner())
}

func TestCallBreak_validatePlay_LeadSpadeBlockedWhenNotBroken(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	cb.players[0].AddCard(NewCard(CardDesignHeart, 3, false))
	cb.currentTrick = nil

	err := cb.validatePlay(0, NewCard(CardDesignSpade, 5, false))
	assert.Error(t, err)
}

func TestCallBreak_validatePlay_LeadSpadeOKWhenBroken(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	cb.players[0].AddCard(NewCard(CardDesignHeart, 3, false))
	cb.currentTrick = nil
	cb.spadesBroken = true

	err := cb.validatePlay(0, NewCard(CardDesignSpade, 5, false))
	assert.NoError(t, err)
}

func TestCallBreak_validatePlay_LeadSpadeOKWhenOnlySpades(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	cb.currentTrick = nil

	err := cb.validatePlay(0, NewCard(CardDesignSpade, 5, false))
	assert.NoError(t, err)
}

func TestCallBreak_validatePlay_MustFollowSuit(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignHeart, 5, false))
	cb.players[0].AddCard(NewCard(CardDesignClover, 3, false))
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 9, false)},
	}

	err := cb.validatePlay(0, NewCard(CardDesignClover, 3, false))
	assert.Error(t, err)

	err = cb.validatePlay(0, NewCard(CardDesignHeart, 5, false))
	assert.NoError(t, err)
}

func TestCallBreak_validatePlay_MustTrumpWhenVoid(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[0].Reset()
	// Heart リードに対してプレイヤー 0 はハート無し、スペードあり、クラブあり
	cb.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	cb.players[0].AddCard(NewCard(CardDesignClover, 3, false))
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 9, false)},
	}

	// クラブを出そうとするとエラー (スペードで切らなければならない)
	err := cb.validatePlay(0, NewCard(CardDesignClover, 3, false))
	assert.Error(t, err)

	// スペードはOK
	err = cb.validatePlay(0, NewCard(CardDesignSpade, 5, false))
	assert.NoError(t, err)
}

func TestCallBreak_validatePlay_DiscardWhenVoidAndNoTrump(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[0].Reset()
	// Heart リードに対してハート無し・スペード無し
	cb.players[0].AddCard(NewCard(CardDesignClover, 3, false))
	cb.players[0].AddCard(NewCard(CardDesignDiamond, 7, false))
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 9, false)},
	}

	err := cb.validatePlay(0, NewCard(CardDesignClover, 3, false))
	assert.NoError(t, err)
	err = cb.validatePlay(0, NewCard(CardDesignDiamond, 7, false))
	assert.NoError(t, err)
}

func TestCallBreak_validatePlay_SpadeLeadFollowSuit(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	cb.players[0].AddCard(NewCard(CardDesignHeart, 3, false))
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 9, false)},
	}
	// スペードリードに対して非スペードを出せない (スペードを持っている)
	err := cb.validatePlay(0, NewCard(CardDesignHeart, 3, false))
	assert.Error(t, err)
}

func TestCallBreak_playCard_MarksSpadesBroken(t *testing.T) {
	cb := newInternalCallBreak()
	cb.SetPhase(CallBreakPhasePlay)
	cb.currentTrick = nil
	cb.SetCurrentPlayerIdx(0)
	cb.playCard(0, NewCard(CardDesignSpade, 5, false))
	assert.True(t, cb.GetSpadesBroken())
}

func TestCallBreak_playCard_AdvancesToTrickEnd(t *testing.T) {
	cb := newInternalCallBreak()
	cb.SetPhase(CallBreakPhasePlay)
	cb.SetCurrentPlayerIdx(3)
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 2, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 3, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 4, false)},
	}
	cb.playCard(3, NewCard(CardDesignHeart, 5, false))
	assert.Equal(t, CallBreakPhaseTrickEnd, cb.GetPhase())
}

func TestCallBreak_cpuBidEasy_Range(t *testing.T) {
	cb := newInternalCallBreak()
	cb.Reset()
	for i := 0; i < 100; i++ {
		bid := cb.cpuBidEasy(1)
		assert.GreaterOrEqual(t, bid, CallBreakMinBid)
		assert.LessOrEqual(t, bid, 5)
	}
}

func TestCallBreak_cpuBidNormal(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[1].Reset()
	cb.players[1].AddCard(NewCard(CardDesignSpade, 12, false))
	cb.players[1].AddCard(NewCard(CardDesignHeart, 13, false))
	cb.players[1].AddCard(NewCard(CardDesignClover, 12, false))
	bid := cb.cpuBidNormal(1)
	assert.GreaterOrEqual(t, bid, CallBreakMinBid)
	assert.LessOrEqual(t, bid, CallBreakHandSize)
}

func TestCallBreak_cpuBidNormal_MinClamp(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[1].Reset()
	for i := 2; i <= 9; i++ {
		cb.players[1].AddCard(NewCard(CardDesignHeart, i, false))
	}
	bid := cb.cpuBidNormal(1)
	assert.Equal(t, CallBreakMinBid, bid, "weak hand should clamp to minimum bid")
}

func TestCallBreak_cpuBidHard(t *testing.T) {
	cb := newInternalCallBreak()
	cb.Reset()
	cb.config.CpuDifficulty = CallBreakCpuDifficultyHard
	// 1000 trials to cover both branches of Q random coin
	for i := 0; i < 100; i++ {
		bid := cb.cpuBidHard(1)
		assert.GreaterOrEqual(t, bid, CallBreakMinBid)
		assert.LessOrEqual(t, bid, CallBreakHandSize)
	}
}

func TestCallBreak_cpuSelectBid_Difficulties(t *testing.T) {
	for _, diff := range []CallBreakCpuDifficulty{CallBreakCpuDifficultyEasy, CallBreakCpuDifficultyNormal, CallBreakCpuDifficultyHard} {
		cb := newInternalCallBreak()
		cb.Reset()
		cb.config.CpuDifficulty = diff
		bid := cb.cpuSelectBid(1)
		assert.GreaterOrEqual(t, bid, CallBreakMinBid)
	}
}

func TestCallBreak_cpuPlayEasy_RandomFromValid(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[1].Reset()
	cb.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
	cb.players[1].AddCard(NewCard(CardDesignHeart, 9, false))
	cb.currentTrick = nil
	idx := cb.cpuPlayEasy([]int{0, 1})
	assert.True(t, idx == 0 || idx == 1)
}

func TestCallBreak_cpuPlayNormal_LeadHighWhenBehind(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[1].Reset()
	cb.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
	cb.players[1].AddCard(NewCard(CardDesignHeart, 11, false))
	cb.players[1].SetBid(3)
	cb.currentTrick = nil
	idx := cb.cpuPlayNormal(1, []int{0, 1})
	assert.Equal(t, 1, idx, "should pick highest when leading and behind on bid")
}

func TestCallBreak_cpuPlayNormal_LeadLowWhenAhead(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[1].Reset()
	cb.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
	cb.players[1].AddCard(NewCard(CardDesignHeart, 11, false))
	cb.players[1].SetBid(1)
	cb.players[1].AddTrick([]*Card{NewCard(CardDesignClover, 2, false)})
	cb.currentTrick = nil
	idx := cb.cpuPlayNormal(1, []int{0, 1})
	assert.Equal(t, 0, idx, "should pick lowest when ahead on bid")
}

func TestCallBreak_cpuPlayHard_LeadStrongWhenBehind(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[1].Reset()
	cb.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
	cb.players[1].AddCard(NewCard(CardDesignSpade, 3, false))
	cb.players[1].SetBid(3)
	cb.currentTrick = nil
	idx := cb.cpuPlayHard(1, []int{0, 1})
	// Spade gets +100 score => should pick the spade
	assert.Equal(t, 1, idx)
}

func TestCallBreak_cpuPlayHard_PicksWinningOver(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[1].Reset()
	cb.players[1].AddCard(NewCard(CardDesignHeart, 4, false))
	cb.players[1].AddCard(NewCard(CardDesignHeart, 12, false))
	cb.players[1].SetBid(3)
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 10, false)},
	}
	idx := cb.cpuPlayHard(1, []int{0, 1})
	assert.Equal(t, 1, idx, "should choose card that wins the trick")
}

func TestCallBreak_cpuPlayHard_DiscardHighWhenNoTrump(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[1].Reset()
	cb.players[1].AddCard(NewCard(CardDesignClover, 4, false))
	cb.players[1].AddCard(NewCard(CardDesignClover, 13, false))
	cb.players[1].SetBid(0) // bid 達成済みのつもり (ahead)
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 10, false)},
	}
	idx := cb.cpuPlayHard(1, []int{0, 1})
	assert.Equal(t, 1, idx)
}

func TestCallBreak_cpuSelectPlayCard_Difficulties(t *testing.T) {
	for _, diff := range []CallBreakCpuDifficulty{CallBreakCpuDifficultyEasy, CallBreakCpuDifficultyNormal, CallBreakCpuDifficultyHard} {
		cb := newInternalCallBreak()
		cb.Reset()
		cb.config.CpuDifficulty = diff
		cb.SetPhase(CallBreakPhasePlay)
		cb.SetCurrentPlayerIdx(1)
		cb.currentTrick = nil
		idx := cb.cpuSelectPlayCard(1)
		assert.GreaterOrEqual(t, idx, 0)
	}
}

func TestCallBreak_cpuSelectPlayCard_SingleValid(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[1].Reset()
	cb.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
	cb.currentTrick = nil
	idx := cb.cpuSelectPlayCard(1)
	assert.Equal(t, 0, idx)
}

func TestCallBreak_cpuSelectPlayCard_NoValid(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[1].Reset() // 手札なし
	cb.currentTrick = nil
	idx := cb.cpuSelectPlayCard(1)
	assert.Equal(t, 0, idx)
}

func TestCallBreak_playHintReason_Variants(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignHeart, 11, false))
	cb.players[0].SetBid(3)
	cb.currentTrick = nil
	reason := cb.playHintReason(0)
	assert.Equal(t, "lead_strong", reason)

	cb.players[0].AddTrick([]*Card{NewCard(CardDesignClover, 2, false)})
	cb.players[0].AddTrick([]*Card{NewCard(CardDesignClover, 3, false)})
	cb.players[0].AddTrick([]*Card{NewCard(CardDesignClover, 4, false)})
	reason = cb.playHintReason(0)
	assert.Equal(t, "lead_low", reason)

	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignHeart, 5, false))
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 11, false)},
	}
	assert.Equal(t, "follow_suit", cb.playHintReason(0))

	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 11, false)},
	}
	assert.Equal(t, "trump_cut", cb.playHintReason(0))

	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignClover, 5, false))
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 11, false)},
	}
	assert.Equal(t, "discard_high", cb.playHintReason(0))
}

func TestCallBreak_PlayerPlay_AllowedWhenValid(t *testing.T) {
	cb := newInternalCallBreak()
	cb.Reset()
	cb.SetPhase(CallBreakPhasePlay)
	cb.SetCurrentPlayerIdx(0)
	cb.currentTrick = nil
	cb.spadesBroken = true

	require.NotZero(t, cb.players[0].GetCardsSize())
	err := cb.PlayerPlay(0)
	assert.NoError(t, err)
}

func TestCallBreak_CpuPlay_AdvancesTurn(t *testing.T) {
	cb := newInternalCallBreak()
	cb.Reset()
	cb.SetPhase(CallBreakPhasePlay)
	cb.SetCurrentPlayerIdx(1)
	cb.currentTrick = nil
	cb.spadesBroken = true

	before := cb.players[1].GetCardsSize()
	cb.CpuPlay()
	assert.Equal(t, before-1, cb.players[1].GetCardsSize())
}

func TestCallBreak_CpuPlay_SkipsHuman(t *testing.T) {
	cb := newInternalCallBreak()
	cb.Reset()
	cb.SetPhase(CallBreakPhasePlay)
	cb.SetCurrentPlayerIdx(0)
	cb.currentTrick = nil
	before := cb.players[0].GetCardsSize()
	cb.CpuPlay()
	assert.Equal(t, before, cb.players[0].GetCardsSize())
}

func TestCallBreak_filterByDesign(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignHeart, 3, false))
	cb.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	cb.players[0].AddCard(NewCard(CardDesignHeart, 7, false))
	out := filterByDesign(cb.players[0], []int{0, 1, 2}, CardDesignHeart)
	assert.Equal(t, []int{0, 2}, out)
}

func TestCallBreak_filterAbove(t *testing.T) {
	cb := newInternalCallBreak()
	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignHeart, 3, false))
	cb.players[0].AddCard(NewCard(CardDesignHeart, 9, false))
	cb.players[0].AddCard(NewCard(CardDesignHeart, 11, false))
	out := filterAbove(cb.players[0], []int{0, 1, 2}, 5)
	assert.Equal(t, []int{1, 2}, out)
}

func TestCallBreak_summariseTrick(t *testing.T) {
	cb := newInternalCallBreak()
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 10, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 5, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 11, false)},
	}
	hs, has, hl := cb.summariseTrick(CardDesignHeart)
	assert.Equal(t, 11, hs)
	assert.True(t, has)
	assert.Equal(t, 10, hl)
}
