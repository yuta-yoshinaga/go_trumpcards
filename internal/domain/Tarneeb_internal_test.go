//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newInternalTarneeb(diff TarneebCpuDifficulty) *Tarneeb {
	players := []*TarneebPlayer{
		NewTarneebPlayer(true, 0),
		NewTarneebPlayer(false, 1),
		NewTarneebPlayer(false, 0),
		NewTarneebPlayer(false, 1),
	}
	cfg := DefaultTarneebConfig()
	cfg.CpuDifficulty = diff
	return NewTarneeb(NewTrumpCards(0), players, cfg)
}

// fillHand: テスト用ヘルパ。プレイヤーの手札をクリアして任意のカードを詰める。
func fillHand(p *TarneebPlayer, cards []*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestCpuSelectTrumpEasy_ReturnsValidSuit(t *testing.T) {
	tn := newInternalTarneeb(TarneebCpuDifficultyEasy)
	suit := tn.cpuSelectTrumpEasy(1)
	assert.True(t, suit >= CardDesignSpade && suit <= CardDesignDiamond)
}

func TestCpuSelectTrumpNormal_LongestSuit(t *testing.T) {
	tn := newInternalTarneeb(TarneebCpuDifficultyNormal)
	fillHand(tn.GetPlayer(1), []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignClover, 4, false),
	})
	assert.Equal(t, CardDesignHeart, tn.cpuSelectTrumpNormal(1))
}

func TestCpuSelectTrumpHard_FavorsHighStrength(t *testing.T) {
	tn := newInternalTarneeb(TarneebCpuDifficultyHard)
	// ♥: 短いが A+K+Q ; ♦: 長いが弱札
	fillHand(tn.GetPlayer(1), []*Card{
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignHeart, 12, false),
		NewCard(CardDesignDiamond, 2, false),
		NewCard(CardDesignDiamond, 3, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignDiamond, 5, false),
	})
	assert.Equal(t, CardDesignHeart, tn.cpuSelectTrumpHard(1))
}

func TestCpuBidEasy_HighCardsTrigger(t *testing.T) {
	tn := newInternalTarneeb(TarneebCpuDifficultyEasy)
	fillHand(tn.GetPlayer(1), []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 7, false),
	})
	// 高札なし → パス
	assert.Equal(t, TarneebPassBid, tn.cpuBidEasy(1))
}

func TestCpuBidNormal_NoHighCards_Passes(t *testing.T) {
	tn := newInternalTarneeb(TarneebCpuDifficultyNormal)
	fillHand(tn.GetPlayer(1), []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignClover, 3, false),
	})
	assert.Equal(t, TarneebPassBid, tn.cpuBidNormal(1))
}

func TestCpuBidHard_WithStrongHand_Bids(t *testing.T) {
	tn := newInternalTarneeb(TarneebCpuDifficultyHard)
	// 6枚スペード (A,K,Q,J,10,9) + 高札3枚 + その他 → 推定 8 トリック相当
	fillHand(tn.GetPlayer(1), []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignHeart, 12, false),
		NewCard(CardDesignDiamond, 1, false),
		NewCard(CardDesignDiamond, 13, false),
	})
	bid := tn.cpuBidHard(1)
	assert.GreaterOrEqual(t, bid, tn.config.MinBid)
}

func TestAdjustToValidBid_ClampsAndPasses(t *testing.T) {
	tn := newInternalTarneeb(TarneebCpuDifficultyNormal)
	tn.SetHighestBid(0)
	assert.Equal(t, tn.config.MinBid, tn.adjustToValidBid(3))
	tn.SetHighestBid(10)
	assert.Equal(t, 11, tn.adjustToValidBid(10))
	tn.SetHighestBid(13)
	assert.Equal(t, TarneebPassBid, tn.adjustToValidBid(13))
}

func TestCpuPlayEasy_PicksFromValid(t *testing.T) {
	tn := newInternalTarneeb(TarneebCpuDifficultyEasy)
	fillHand(tn.GetPlayer(1), []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 3, false),
	})
	idx := tn.cpuPlayEasy([]int{0, 1})
	assert.True(t, idx == 0 || idx == 1)
}

func TestCpuPlayNormal_LeadHighest(t *testing.T) {
	tn := newInternalTarneeb(TarneebCpuDifficultyNormal)
	tn.SetPhase(TarneebPhasePlay)
	tn.SetTrumpSuit(CardDesignSpade)
	fillHand(tn.GetPlayer(1), []*Card{
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignHeart, 13, false),
	})
	tn.SetCurrentTrick(nil)
	idx := tn.cpuPlayNormal(1, []int{0, 1})
	assert.Equal(t, 1, idx)
}

func TestCpuPlayHard_PartnerWinning_PlaysLow(t *testing.T) {
	tn := newInternalTarneeb(TarneebCpuDifficultyHard)
	tn.SetPhase(TarneebPhasePlay)
	tn.SetTrumpSuit(CardDesignSpade)
	// idx 2 (team 0) leads with A♥; idx 1 (team 1) is current player; idx 3 (team 1) is partner of idx 1.
	// idx 3 hasn't played yet, but trick winner consult is for idx 1's choice with idx 0 (team 0) being current trick.
	// パートナーが勝っている状況を模擬: idx 3 がパートナーで A♥ をリード。
	tn.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 1, false)}, // partner of idx1
	})
	fillHand(tn.GetPlayer(1), []*Card{
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignHeart, 10, false),
	})
	tn.SetCurrentPlayerIdx(1)
	idx := tn.cpuPlayHard(1, []int{0, 1})
	// パートナー勝ち → 低いカード
	assert.Equal(t, 0, idx)
}

func TestSuitName(t *testing.T) {
	assert.Equal(t, "♠", suitName(CardDesignSpade))
	assert.Equal(t, "♣", suitName(CardDesignClover))
	assert.Equal(t, "♥", suitName(CardDesignHeart))
	assert.Equal(t, "♦", suitName(CardDesignDiamond))
	assert.Equal(t, "?", suitName(99))
}

func TestIsValidSuit(t *testing.T) {
	assert.True(t, isValidSuit(CardDesignSpade))
	assert.True(t, isValidSuit(CardDesignDiamond))
	assert.False(t, isValidSuit(0))
	assert.False(t, isValidSuit(99))
}

func TestSummariseTrick(t *testing.T) {
	tn := newInternalTarneeb(TarneebCpuDifficultyNormal)
	tn.SetTrumpSuit(CardDesignSpade)
	tn.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 5, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 2, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 13, false)},
	})
	maxLead, hasLead, maxTrump, hasTrump := tn.summariseTrick(CardDesignHeart)
	assert.Equal(t, 13, maxLead)
	assert.True(t, hasLead)
	assert.Equal(t, 2, maxTrump)
	assert.True(t, hasTrump)
}

func TestHintBidReason(t *testing.T) {
	assert.Equal(t, "bid_pass", hintBidReason(TarneebPassBid))
	assert.Equal(t, "bid_estimate", hintBidReason(8))
}

func TestPlayHintReason(t *testing.T) {
	tn := newInternalTarneeb(TarneebCpuDifficultyNormal)
	tn.SetTrumpSuit(CardDesignSpade)
	fillHand(tn.GetPlayer(0), []*Card{
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignDiamond, 10, false),
	})
	tn.SetCurrentTrick(nil)
	assert.Equal(t, "lead_strong", tn.playHintReason(0, 0))

	tn.SetCurrentTrick([]*TrickCard{
		{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 8, false)},
	})
	assert.Equal(t, "follow_suit", tn.playHintReason(0, 0))
	assert.Equal(t, "trump_cut", tn.playHintReason(0, 1))
	assert.Equal(t, "discard_high", tn.playHintReason(0, 2))
}
