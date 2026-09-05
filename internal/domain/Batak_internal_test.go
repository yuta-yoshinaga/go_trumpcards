//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newInternalBatak() *Batak {
	players := []*BatakPlayer{
		NewBatakPlayer(true),
		NewBatakPlayer(false),
		NewBatakPlayer(false),
		NewBatakPlayer(false),
	}
	return NewBatak(NewTrumpCards(0), players, DefaultBatakConfig())
}

func TestBatak_findHumanIdx(t *testing.T) {
	cb := newInternalBatak()
	assert.Equal(t, 0, findHumanIdx(cb.players))

	allCpu := []*BatakPlayer{
		NewBatakPlayer(false),
		NewBatakPlayer(false),
		NewBatakPlayer(false),
		NewBatakPlayer(false),
	}
	cb2 := NewBatak(NewTrumpCards(0), allCpu, DefaultBatakConfig())
	assert.Equal(t, -1, findHumanIdx(cb2.players))
}

func TestBatak_playerHasSuit(t *testing.T) {
	cb := newInternalBatak()
	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	assert.True(t, cb.playerHasSuit(0, CardDesignSpade))
	assert.False(t, cb.playerHasSuit(0, CardDesignHeart))
}

func TestBatak_playerHasNonSpade(t *testing.T) {
	cb := newInternalBatak()
	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	assert.False(t, cb.playerHasNonSpade(0))
	cb.players[0].AddCard(NewCard(CardDesignHeart, 3, false))
	assert.True(t, cb.playerHasNonSpade(0))
}

func TestBatak_trickWinner_EmptyTrick(t *testing.T) {
	cb := newInternalBatak()
	cb.currentTrick = nil
	assert.Equal(t, 0, cb.trickWinner())
}

func TestBatak_trickWinner_LeadSuitWins(t *testing.T) {
	cb := newInternalBatak()
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 5, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 11, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 4, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignClover, 13, false)},
	}
	assert.Equal(t, 1, cb.trickWinner())
}

func TestBatak_trickWinner_SpadeBeatsLeadSuit(t *testing.T) {
	cb := newInternalBatak()
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 13, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 2, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 12, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 11, false)},
	}
	assert.Equal(t, 1, cb.trickWinner())
}

func TestBatak_trickWinner_HigherSpadeWins(t *testing.T) {
	cb := newInternalBatak()
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 13, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 2, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 7, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 5, false)},
	}
	assert.Equal(t, 2, cb.trickWinner())
}

func TestBatak_trickWinner_AllSpades(t *testing.T) {
	cb := newInternalBatak()
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 3, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 11, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 7, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)},
	}
	assert.Equal(t, 3, cb.trickWinner())
}

func TestBatak_validatePlay_LeadSpadeBlockedWhenNotBroken(t *testing.T) {
	cb := newInternalBatak()
	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	cb.players[0].AddCard(NewCard(CardDesignHeart, 3, false))
	cb.currentTrick = nil

	err := cb.validatePlay(0, NewCard(CardDesignSpade, 5, false))
	assert.Error(t, err)
}

func TestBatak_validatePlay_LeadSpadeOKWhenBroken(t *testing.T) {
	cb := newInternalBatak()
	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	cb.players[0].AddCard(NewCard(CardDesignHeart, 3, false))
	cb.currentTrick = nil
	cb.spadesBroken = true

	err := cb.validatePlay(0, NewCard(CardDesignSpade, 5, false))
	assert.NoError(t, err)
}

func TestBatak_validatePlay_LeadSpadeOKWhenOnlySpades(t *testing.T) {
	cb := newInternalBatak()
	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	cb.currentTrick = nil

	err := cb.validatePlay(0, NewCard(CardDesignSpade, 5, false))
	assert.NoError(t, err)
}

func TestBatak_validatePlay_MustFollowSuit(t *testing.T) {
	cb := newInternalBatak()
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

func TestBatak_validatePlay_MustTrumpWhenVoid(t *testing.T) {
	cb := newInternalBatak()
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

func TestBatak_validatePlay_DiscardWhenVoidAndNoTrump(t *testing.T) {
	cb := newInternalBatak()
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

func TestBatak_validatePlay_SpadeLeadFollowSuit(t *testing.T) {
	cb := newInternalBatak()
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

func TestBatak_playCard_MarksSpadesBroken(t *testing.T) {
	cb := newInternalBatak()
	cb.SetPhase(BatakPhasePlay)
	cb.currentTrick = nil
	cb.SetCurrentPlayerIdx(0)
	cb.playCard(0, NewCard(CardDesignSpade, 5, false))
	assert.True(t, cb.GetSpadesBroken())
}

func TestBatak_playCard_AdvancesToTrickEnd(t *testing.T) {
	cb := newInternalBatak()
	cb.SetPhase(BatakPhasePlay)
	cb.SetCurrentPlayerIdx(3)
	cb.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 2, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 3, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 4, false)},
	}
	cb.playCard(3, NewCard(CardDesignHeart, 5, false))
	assert.Equal(t, BatakPhaseTrickEnd, cb.GetPhase())
}

func makeBatakTestHands() (strongHand []*Card, weakHand []*Card) {
	// 強い手札: スペード 5 枚 (A, K, Q, 10, 9), ハート A, K, ダイヤ A, 5, クラブ 2, 3, 4, 5 (計13枚)
	strongHand = []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignDiamond, 1, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignClover, 4, false),
		NewCard(CardDesignClover, 5, false),
	}
	// 弱い手札: スペードなし・絵札なし (計13枚)
	weakHand = []*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignHeart, 6, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignHeart, 8, false),
		NewCard(CardDesignDiamond, 2, false),
		NewCard(CardDesignDiamond, 3, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignClover, 4, false),
	}
	return
}

func TestBatak_cpuBidEasy_Range(t *testing.T) {
	cb := newInternalBatak()
	cb.Reset()

	strongHand, weakHand := makeBatakTestHands()

	// 弱い手札ではパス (0) を返すこと
	cb.players[1].Reset()
	for _, c := range weakHand {
		cb.players[1].AddCard(c)
	}
	assert.Equal(t, BatakPassBid, cb.cpuBidEasy(1), "弱い手札では Easy でもパスすること")

	// 強い手札では 5 以上のビッドを返すこと
	cb.players[1].Reset()
	for _, c := range strongHand {
		cb.players[1].AddCard(c)
	}
	for i := 0; i < 50; i++ {
		bid := cb.cpuBidEasy(1)
		assert.GreaterOrEqual(t, bid, BatakMinBid, "強い手札では BatakMinBid (5) 以上であること")
		assert.LessOrEqual(t, bid, BatakMinBid+2, "Easy のビッド上限以下であること")
	}

	// ランダム配りでの範囲チェック
	cb.Reset()
	for i := 0; i < 100; i++ {
		bid := cb.cpuBidEasy(1)
		assert.GreaterOrEqual(t, bid, 0)
		assert.LessOrEqual(t, bid, BatakHandSize)
	}
}

func TestBatak_cpuBidNormal(t *testing.T) {
	cb := newInternalBatak()
	cb.Reset()

	strongHand, weakHand := makeBatakTestHands()

	// 強い手札では 5 以上を見積もること
	cb.players[1].Reset()
	for _, c := range strongHand {
		cb.players[1].AddCard(c)
	}
	bidStrong := cb.cpuBidNormal(1)
	assert.GreaterOrEqual(t, bidStrong, 5, "強い手札は5以上を見積もること: got %d", bidStrong)

	// 弱い手札では 5 未満 (0) を見積もること
	cb.players[1].Reset()
	for _, c := range weakHand {
		cb.players[1].AddCard(c)
	}
	bidWeak := cb.cpuBidNormal(1)
	assert.Less(t, bidWeak, 5, "弱い手札は5未満を見積もること: got %d", bidWeak)
	assert.Equal(t, 0, bidWeak)
}

func TestBatak_cpuBidNormal_WeakHand(t *testing.T) {
	cb := newInternalBatak()
	cb.players[1].Reset()
	for i := 2; i <= 9; i++ {
		cb.players[1].AddCard(NewCard(CardDesignHeart, i, false))
	}
	bid := cb.cpuBidNormal(1)
	assert.Equal(t, 0, bid, "weak hand should evaluate to 0")
}

func TestBatak_clampBid(t *testing.T) {
	cb := newInternalBatak()
	assert.Equal(t, 0, cb.clampBid(0))
	assert.Equal(t, 4, cb.clampBid(4))
	assert.Equal(t, 5, cb.clampBid(5))
	assert.Equal(t, 13, cb.clampBid(13))
	assert.Equal(t, 13, cb.clampBid(14))
	assert.Equal(t, 13, cb.clampBid(100))
}

func TestBatak_cpuBidHard(t *testing.T) {
	cb := newInternalBatak()
	cb.Reset()
	cb.config.CpuDifficulty = BatakCpuDifficultyHard

	strongHand, weakHand := makeBatakTestHands()

	// 強い手札では 5 以上を見積もること
	cb.players[1].Reset()
	for _, c := range strongHand {
		cb.players[1].AddCard(c)
	}
	bidStrong := cb.cpuBidHard(1)
	assert.GreaterOrEqual(t, bidStrong, 5, "強い手札は5以上を見積もること: got %d", bidStrong)

	// 弱い手札では 5 未満 (0) を見積もること
	cb.players[1].Reset()
	for _, c := range weakHand {
		cb.players[1].AddCard(c)
	}
	bidWeak := cb.cpuBidHard(1)
	assert.Less(t, bidWeak, 5, "弱い手札は5未満を見積もること: got %d", bidWeak)
	assert.Equal(t, 0, bidWeak)

	// 100 配りで範囲チェック
	for i := 0; i < 100; i++ {
		cb.Reset()
		bid := cb.cpuBidHard(1)
		assert.GreaterOrEqual(t, bid, 0)
		assert.LessOrEqual(t, bid, BatakHandSize)
	}
}

func TestBatak_cpuSelectBid_Difficulties(t *testing.T) {
	for _, diff := range []BatakCpuDifficulty{BatakCpuDifficultyEasy, BatakCpuDifficultyNormal, BatakCpuDifficultyHard} {
		cb := newInternalBatak()
		cb.Reset()
		cb.config.CpuDifficulty = diff
		bid := cb.cpuSelectBid(1)
		assert.GreaterOrEqual(t, bid, 0)
		assert.LessOrEqual(t, bid, BatakHandSize)
	}
}

func TestBatak_cpuPlayEasy_RandomFromValid(t *testing.T) {
	cb := newInternalBatak()
	cb.players[1].Reset()
	cb.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
	cb.players[1].AddCard(NewCard(CardDesignHeart, 9, false))
	cb.currentTrick = nil
	idx := cb.cpuPlayEasy([]int{0, 1})
	assert.True(t, idx == 0 || idx == 1)
}

func TestBatak_cpuPlayNormal_LeadHighWhenBehind(t *testing.T) {
	cb := newInternalBatak()
	cb.players[1].Reset()
	cb.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
	cb.players[1].AddCard(NewCard(CardDesignHeart, 11, false))
	cb.players[1].SetBid(3)
	cb.currentTrick = nil
	idx := cb.cpuPlayNormal(1, []int{0, 1})
	assert.Equal(t, 1, idx, "should pick highest when leading and behind on bid")
}

func TestBatak_cpuPlayNormal_LeadLowWhenAhead(t *testing.T) {
	cb := newInternalBatak()
	cb.players[1].Reset()
	cb.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
	cb.players[1].AddCard(NewCard(CardDesignHeart, 11, false))
	cb.players[1].SetBid(1)
	cb.players[1].AddTrick([]*Card{NewCard(CardDesignClover, 2, false)})
	cb.currentTrick = nil
	idx := cb.cpuPlayNormal(1, []int{0, 1})
	assert.Equal(t, 0, idx, "should pick lowest when ahead on bid")
}

func TestBatak_cpuPlayHard_LeadStrongWhenBehind(t *testing.T) {
	cb := newInternalBatak()
	cb.players[1].Reset()
	cb.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
	cb.players[1].AddCard(NewCard(CardDesignSpade, 3, false))
	cb.players[1].SetBid(3)
	cb.currentTrick = nil
	idx := cb.cpuPlayHard(1, []int{0, 1})
	// Spade gets +100 score => should pick the spade
	assert.Equal(t, 1, idx)
}

func TestBatak_cpuPlayHard_PicksWinningOver(t *testing.T) {
	cb := newInternalBatak()
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

func TestBatak_cpuPlayHard_DiscardHighWhenNoTrump(t *testing.T) {
	cb := newInternalBatak()
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

func TestBatak_cpuSelectPlayCard_Difficulties(t *testing.T) {
	for _, diff := range []BatakCpuDifficulty{BatakCpuDifficultyEasy, BatakCpuDifficultyNormal, BatakCpuDifficultyHard} {
		cb := newInternalBatak()
		cb.Reset()
		cb.config.CpuDifficulty = diff
		cb.SetPhase(BatakPhasePlay)
		cb.SetCurrentPlayerIdx(1)
		cb.currentTrick = nil
		idx := cb.cpuSelectPlayCard(1)
		assert.GreaterOrEqual(t, idx, 0)
	}
}

func TestBatak_cpuSelectPlayCard_SingleValid(t *testing.T) {
	cb := newInternalBatak()
	cb.players[1].Reset()
	cb.players[1].AddCard(NewCard(CardDesignHeart, 5, false))
	cb.currentTrick = nil
	idx := cb.cpuSelectPlayCard(1)
	assert.Equal(t, 0, idx)
}

func TestBatak_cpuSelectPlayCard_NoValid(t *testing.T) {
	cb := newInternalBatak()
	cb.players[1].Reset() // 手札なし
	cb.currentTrick = nil
	idx := cb.cpuSelectPlayCard(1)
	assert.Equal(t, 0, idx)
}

func TestBatak_playHintReason_Variants(t *testing.T) {
	cb := newInternalBatak()
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

func TestBatak_PlayerPlay_AllowedWhenValid(t *testing.T) {
	cb := newInternalBatak()
	cb.Reset()
	cb.SetPhase(BatakPhasePlay)
	cb.SetCurrentPlayerIdx(0)
	cb.currentTrick = nil
	cb.spadesBroken = true

	require.NotZero(t, cb.players[0].GetCardsSize())
	err := cb.PlayerPlay(0)
	assert.NoError(t, err)
}

func TestBatak_CpuPlay_AdvancesTurn(t *testing.T) {
	cb := newInternalBatak()
	cb.Reset()
	cb.SetPhase(BatakPhasePlay)
	cb.SetCurrentPlayerIdx(1)
	cb.currentTrick = nil
	cb.spadesBroken = true

	before := cb.players[1].GetCardsSize()
	cb.CpuPlay()
	assert.Equal(t, before-1, cb.players[1].GetCardsSize())
}

func TestBatak_CpuPlay_SkipsHuman(t *testing.T) {
	cb := newInternalBatak()
	cb.Reset()
	cb.SetPhase(BatakPhasePlay)
	cb.SetCurrentPlayerIdx(0)
	cb.currentTrick = nil
	before := cb.players[0].GetCardsSize()
	cb.CpuPlay()
	assert.Equal(t, before, cb.players[0].GetCardsSize())
}

func TestBatak_filterByDesign(t *testing.T) {
	cb := newInternalBatak()
	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignHeart, 3, false))
	cb.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	cb.players[0].AddCard(NewCard(CardDesignHeart, 7, false))
	out := filterByDesign(cb.players[0], []int{0, 1, 2}, CardDesignHeart)
	assert.Equal(t, []int{0, 2}, out)
}

func TestBatak_filterAbove(t *testing.T) {
	cb := newInternalBatak()
	cb.players[0].Reset()
	cb.players[0].AddCard(NewCard(CardDesignHeart, 3, false))
	cb.players[0].AddCard(NewCard(CardDesignHeart, 9, false))
	cb.players[0].AddCard(NewCard(CardDesignHeart, 11, false))
	out := filterAbove(cb.players[0], []int{0, 1, 2}, 5, nil)
	assert.Equal(t, []int{1, 2}, out)
}

func TestBatak_summariseTrick(t *testing.T) {
	cb := newInternalBatak()
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

// TestBatak_BidEstimation_Statistics は 2000 配りで cpuBidNormal / cpuBidHard の
// 見積り平均および宣言可能手 (>=5) の割合を統計的に検証する恒久的テスト。
func TestBatak_BidEstimation_Statistics(t *testing.T) {
	cb := newInternalBatak()
	trials := 2000

	normTotal := 0
	hardTotal := 0
	normGte5 := 0
	hardGte5 := 0
	minNorm, maxNorm := BatakHandSize, 0
	minHard, maxHard := BatakHandSize, 0

	for i := 0; i < trials; i++ {
		cb.Reset()
		bNorm := cb.cpuBidNormal(0)
		bHard := cb.cpuBidHard(0)
		normTotal += bNorm
		hardTotal += bHard
		if bNorm >= 5 {
			normGte5++
		}
		if bHard >= 5 {
			hardGte5++
		}
		if bNorm < minNorm {
			minNorm = bNorm
		}
		if bNorm > maxNorm {
			maxNorm = bNorm
		}
		if bHard < minHard {
			minHard = bHard
		}
		if bHard > maxHard {
			maxHard = bHard
		}
	}

	normAvg := float64(normTotal) / float64(trials)
	hardAvg := float64(hardTotal) / float64(trials)
	normGte5Pct := float64(normGte5) / float64(trials) * 100
	hardGte5Pct := float64(hardGte5) / float64(trials) * 100

	t.Logf("BidEstimation: Normal avg=%.3f, gte5=%.1f%%; Hard avg=%.3f, gte5=%.1f%%",
		normAvg, normGte5Pct, hardAvg, hardGte5Pct)

	// 1. 見積りの平均: cpuBidNormal / cpuBidHard の平均が 3.0 〜 3.6 に入ること (真の平均 3.25)
	assert.GreaterOrEqual(t, normAvg, 3.0, "Normal 見積り平均は 3.0 以上であること")
	assert.LessOrEqual(t, normAvg, 3.6, "Normal 見積り平均は 3.6 以下であること")
	assert.GreaterOrEqual(t, hardAvg, 3.0, "Hard 見積り平均は 3.0 以上であること")
	assert.LessOrEqual(t, hardAvg, 3.6, "Hard 見積り平均は 3.6 以下であること")

	// 0 や 13 に張り付いていないこと
	assert.Less(t, minNorm, 3, "Normal 見積りの最小値が 3 未満であること (張り付き防止)")
	assert.Greater(t, maxNorm, 5, "Normal 見積りの最大値が 5 超であること (張り付き防止)")
	assert.Less(t, minHard, 3, "Hard 見積りの最小値が 3 未満であること (張り付き防止)")
	assert.Greater(t, maxHard, 5, "Hard 見積りの最大値が 5 超であること (張り付き防止)")

	// 2. 宣言可能な手の割合: 見積り 5 以上が 15% 〜 40%
	assert.GreaterOrEqual(t, normGte5Pct, 15.0, "Normal 宣言可能手割合は 15%% 以上であること")
	assert.LessOrEqual(t, normGte5Pct, 40.0, "Normal 宣言可能手割合は 40%% 以下であること")
	assert.GreaterOrEqual(t, hardGte5Pct, 15.0, "Hard 宣言可能手割合は 15%% 以上であること")
	assert.LessOrEqual(t, hardGte5Pct, 40.0, "Hard 宣言可能手割合は 40%% 以下であること")
}
