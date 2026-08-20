//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tcNorm(value, design int) *Card { return NewCard(design, value, true) }
func tcMahjong() *Card               { return NewCard(CardDesignJoker, TichuMahjong, true) }
func tcDog() *Card                   { return NewCard(CardDesignJoker, TichuDog, true) }
func tcPhoenix() *Card               { return NewCard(CardDesignJoker, TichuPhoenix, true) }
func tcDragon() *Card                { return NewCard(CardDesignJoker, TichuDragon, true) }

func TestTichuCardPoints(t *testing.T) {
	cases := []struct {
		card *Card
		want int
	}{
		{tcNorm(5, CardDesignSpade), 5},
		{tcNorm(10, CardDesignHeart), 10},
		{tcNorm(13, CardDesignClover), 10},
		{tcNorm(7, CardDesignSpade), 0},
		{tcDragon(), 25},
		{tcPhoenix(), -25},
		{tcMahjong(), 0},
		{tcDog(), 0},
	}
	for _, c := range cases {
		if got := TichuCardPoints(c.card); got != c.want {
			t.Errorf("TichuCardPoints = %d, want %d", got, c.want)
		}
	}
	// 通常デッキの合計得点は100
	total := 0
	for d := CardDesignSpade; d <= CardDesignDiamond; d++ {
		for v := 1; v <= 13; v++ {
			total += TichuCardPoints(tcNorm(v, d))
		}
	}
	total += TichuCardPoints(tcDragon()) + TichuCardPoints(tcPhoenix())
	if total != 100 {
		t.Errorf("deck total points = %d, want 100", total)
	}
}

func TestTichuCardStrengthOrder(t *testing.T) {
	if TichuCardStrength(tcMahjong()) != 1 {
		t.Error("mahjong strength should be 1")
	}
	if TichuCardStrength(tcNorm(1, CardDesignSpade)) <= TichuCardStrength(tcNorm(13, CardDesignSpade)) {
		t.Error("ace should outrank king")
	}
	if TichuCardStrength(tcDragon()) <= TichuCardStrength(tcNorm(1, CardDesignSpade)) {
		t.Error("dragon should outrank ace")
	}
}

func TestClassifyTichuBasics(t *testing.T) {
	if c := ClassifyTichu(nil); c != nil {
		t.Error("empty should classify nil")
	}
	if c := ClassifyTichu([]*Card{tcNorm(7, CardDesignSpade)}); c == nil || c.Type != TichuComboSingle || c.Rank != 7 {
		t.Errorf("single 7 misclassified: %+v", c)
	}
	if c := ClassifyTichu([]*Card{tcDragon()}); c == nil || c.Type != TichuComboSingle || c.Rank != tichuDragonRank {
		t.Errorf("dragon single misclassified: %+v", c)
	}
	if c := ClassifyTichu([]*Card{tcDog()}); c == nil || c.Type != TichuComboDog {
		t.Errorf("dog misclassified: %+v", c)
	}
	if c := ClassifyTichu([]*Card{tcPhoenix()}); c == nil || !c.PhoenixSingle {
		t.Errorf("phoenix single misclassified: %+v", c)
	}
	// 犬・龍は複数枚で無効
	if c := ClassifyTichu([]*Card{tcDog(), tcNorm(5, CardDesignSpade)}); c != nil {
		t.Error("dog with extra card should be invalid")
	}
	if c := ClassifyTichu([]*Card{tcDragon(), tcNorm(5, CardDesignSpade)}); c != nil {
		t.Error("dragon with extra card should be invalid")
	}
}

func TestClassifyTichuPairTripleFullHouse(t *testing.T) {
	pair := ClassifyTichu([]*Card{tcNorm(8, CardDesignSpade), tcNorm(8, CardDesignHeart)})
	if pair == nil || pair.Type != TichuComboPair || pair.Rank != 8 {
		t.Errorf("pair misclassified: %+v", pair)
	}
	// 鳳凰ペア
	php := ClassifyTichu([]*Card{tcNorm(13, CardDesignSpade), tcPhoenix()})
	if php == nil || php.Type != TichuComboPair || php.Rank != 13 {
		t.Errorf("phoenix pair misclassified: %+v", php)
	}
	trip := ClassifyTichu([]*Card{tcNorm(9, CardDesignSpade), tcNorm(9, CardDesignHeart), tcNorm(9, CardDesignClover)})
	if trip == nil || trip.Type != TichuComboTriple || trip.Rank != 9 {
		t.Errorf("triple misclassified: %+v", trip)
	}
	fh := ClassifyTichu([]*Card{
		tcNorm(7, CardDesignSpade), tcNorm(7, CardDesignHeart), tcNorm(7, CardDesignClover),
		tcNorm(9, CardDesignSpade), tcNorm(9, CardDesignHeart),
	})
	if fh == nil || fh.Type != TichuComboFullHouse || fh.Rank != 7 {
		t.Errorf("fullhouse misclassified: %+v", fh)
	}
	// 鳳凰フルハウス (2+2 → 高い方がトリップ)
	fhp := ClassifyTichu([]*Card{
		tcNorm(7, CardDesignSpade), tcNorm(7, CardDesignHeart),
		tcNorm(9, CardDesignSpade), tcNorm(9, CardDesignHeart), tcPhoenix(),
	})
	if fhp == nil || fhp.Type != TichuComboFullHouse || fhp.Rank != 9 {
		t.Errorf("phoenix fullhouse misclassified: %+v", fhp)
	}
}

func TestClassifyTichuStraightStairsBomb(t *testing.T) {
	st := ClassifyTichu([]*Card{
		tcNorm(3, CardDesignSpade), tcNorm(4, CardDesignHeart), tcNorm(5, CardDesignClover),
		tcNorm(6, CardDesignSpade), tcNorm(7, CardDesignHeart),
	})
	if st == nil || st.Type != TichuComboStraight || st.Rank != 7 || st.Length != 5 {
		t.Errorf("straight misclassified: %+v", st)
	}
	// 鳳凰で穴埋め
	stp := ClassifyTichu([]*Card{
		tcNorm(3, CardDesignSpade), tcNorm(4, CardDesignHeart), tcPhoenix(),
		tcNorm(6, CardDesignSpade), tcNorm(7, CardDesignHeart),
	})
	if stp == nil || stp.Type != TichuComboStraight || stp.Rank != 7 {
		t.Errorf("phoenix straight misclassified: %+v", stp)
	}
	stairs := ClassifyTichu([]*Card{
		tcNorm(5, CardDesignSpade), tcNorm(5, CardDesignHeart),
		tcNorm(6, CardDesignSpade), tcNorm(6, CardDesignHeart),
	})
	if stairs == nil || stairs.Type != TichuComboStairs || stairs.Rank != 6 || stairs.Length != 4 {
		t.Errorf("stairs misclassified: %+v", stairs)
	}
	bomb := ClassifyTichu([]*Card{
		tcNorm(8, CardDesignSpade), tcNorm(8, CardDesignHeart),
		tcNorm(8, CardDesignClover), tcNorm(8, CardDesignDiamond),
	})
	if bomb == nil || bomb.Type != TichuComboBomb || bomb.Rank != 8 {
		t.Errorf("bomb misclassified: %+v", bomb)
	}
	sf := ClassifyTichu([]*Card{
		tcNorm(3, CardDesignSpade), tcNorm(4, CardDesignSpade), tcNorm(5, CardDesignSpade),
		tcNorm(6, CardDesignSpade), tcNorm(7, CardDesignSpade),
	})
	if sf == nil || sf.Type != TichuComboStraightFlush || sf.Rank != 7 {
		t.Errorf("straight flush misclassified: %+v", sf)
	}
}

func TestTichuCanBeatSingles(t *testing.T) {
	low := ClassifyTichu([]*Card{tcNorm(7, CardDesignSpade)})
	high := ClassifyTichu([]*Card{tcNorm(9, CardDesignSpade)})
	if !TichuCanBeat(high, low) {
		t.Error("9 should beat 7")
	}
	if TichuCanBeat(low, high) {
		t.Error("7 should not beat 9")
	}
	dragon := ClassifyTichu([]*Card{tcDragon()})
	ace := ClassifyTichu([]*Card{tcNorm(1, CardDesignSpade)})
	if !TichuCanBeat(dragon, ace) {
		t.Error("dragon should beat ace")
	}
	// 鳳凰はテーブル依存
	phoenix := ClassifyTichu([]*Card{tcPhoenix()})
	phoenix.Rank = high.Rank // 9 の上
	if !TichuCanBeat(phoenix, high) {
		t.Error("phoenix should beat 9")
	}
	// 鳳凰は龍を超えられない
	phoenix2 := ClassifyTichu([]*Card{tcPhoenix()})
	phoenix2.Rank = dragon.Rank
	if TichuCanBeat(phoenix2, dragon) {
		t.Error("phoenix must not beat dragon")
	}
}

func TestTichuCanBeatBombs(t *testing.T) {
	pair := ClassifyTichu([]*Card{tcNorm(13, CardDesignSpade), tcNorm(13, CardDesignHeart)})
	bomb := ClassifyTichu([]*Card{
		tcNorm(3, CardDesignSpade), tcNorm(3, CardDesignHeart),
		tcNorm(3, CardDesignClover), tcNorm(3, CardDesignDiamond),
	})
	if !TichuCanBeat(bomb, pair) {
		t.Error("bomb should beat a pair")
	}
	if TichuCanBeat(pair, bomb) {
		t.Error("pair should not beat a bomb")
	}
	sf := ClassifyTichu([]*Card{
		tcNorm(3, CardDesignSpade), tcNorm(4, CardDesignSpade), tcNorm(5, CardDesignSpade),
		tcNorm(6, CardDesignSpade), tcNorm(7, CardDesignSpade),
	})
	if !TichuCanBeat(sf, bomb) {
		t.Error("straight flush should beat a four-bomb")
	}
	if TichuCanBeat(bomb, sf) {
		t.Error("four-bomb should not beat a straight flush")
	}
}

func TestTichuCanBeatTypeMismatch(t *testing.T) {
	pair := ClassifyTichu([]*Card{tcNorm(13, CardDesignSpade), tcNorm(13, CardDesignHeart)})
	single := ClassifyTichu([]*Card{tcNorm(14-13+1, CardDesignSpade)})
	if TichuCanBeat(single, pair) {
		t.Error("single must not beat a pair")
	}
	st5 := ClassifyTichu([]*Card{
		tcNorm(3, CardDesignSpade), tcNorm(4, CardDesignHeart), tcNorm(5, CardDesignClover),
		tcNorm(6, CardDesignSpade), tcNorm(7, CardDesignHeart),
	})
	st6 := ClassifyTichu([]*Card{
		tcNorm(3, CardDesignSpade), tcNorm(4, CardDesignHeart), tcNorm(5, CardDesignClover),
		tcNorm(6, CardDesignSpade), tcNorm(7, CardDesignHeart), tcNorm(8, CardDesignClover),
	})
	if TichuCanBeat(st6, st5) {
		t.Error("straights of different length must not compare")
	}
}

// #5635: Web はボムを構成する札に赤いリングと💣を出しているのに、CUI は無印の
// 一覧だけで、手札を目視で数えるしかなかった。
func TestTichuBombIndicesFindsFourOfAKind(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignClover, 7, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignSpade, 2, false),
	}
	assert.Equal(t, []int{0, 1, 2, 3}, TichuBombIndices(cards))
}

func TestTichuBombIndicesFindsAStraightFlush(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignHeart, 6, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignSpade, 9, false),
	}
	assert.Equal(t, []int{0, 1, 2, 3, 4}, TichuBombIndices(cards))
}

// 4枚に届かない・5枚に届かない手札では 1 枚も印を付けない。
func TestTichuBombIndicesFindsNothingWithoutABomb(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignClover, 7, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignHeart, 6, false),
	}
	assert.Empty(t, TichuBombIndices(cards))
}

// **同スート5枚でも連続していなければボムではない。**数え上げ側で長さだけ見て
// 印を付けると、出せない組を「ボム」と呼ぶことになる。
func TestTichuBombIndicesRejectsAGappedFlush(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignHeart, 6, false),
		NewCard(CardDesignHeart, 9, false), // 7,8 が無いので連続しない
	}
	assert.Empty(t, TichuBombIndices(cards))
}

// **数え上げた組は、実際に評価器がボムとして受け取ること。**ここがずれると、
// 画面が「ボム」と言った組を出そうとして弾かれる。
func TestTichuBombIndicesAgreeWithTheEvaluator(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignSpade, 4, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 6, false),
		NewCard(CardDesignSpade, 7, false),
	}
	idx := TichuBombIndices(cards)
	require.NotEmpty(t, idx)

	group := make([]*Card, 0, len(idx))
	for _, i := range idx[:4] {
		group = append(group, cards[i])
	}
	assert.NotNil(t, tichuTryBomb4(group, 0), "4枚組が評価器に通る")

	sf := make([]*Card, 0, 5)
	for _, i := range idx[4:] {
		sf = append(sf, cards[i])
	}
	assert.NotNil(t, tichuTryStraightFlush(sf, 0), "ストレートフラッシュが評価器に通る")
}

// 特殊カード (龍・鳳凰・麻雀・犬) はボムに参加しない。
func TestTichuBombIndicesIgnoresSpecials(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignJoker, 1, false),
		NewCard(CardDesignJoker, 2, false),
		NewCard(CardDesignJoker, 3, false),
		NewCard(CardDesignJoker, 4, false),
	}
	assert.Empty(t, TichuBombIndices(cards))
}
