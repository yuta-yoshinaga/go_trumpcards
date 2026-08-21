//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Big Ben's five divergences from Grandfather's Clock ---
//
// Grandfather's Clock is the clone source: twelve clock foundations with fixed
// starters and per-position target ranks, eight columns of five, and a K-to-A
// wrap. Everything below is a rule it does NOT have, so each test carries a
// negative control that its predicate would fail.

// **2 組 104 枚。**文字盤 12 + タブロー 40 = 52 で、残り 52 が山札。
// クローン元は 1 組 52 枚で山札を持たない。
func TestBigBen_UsesTwoDecksAndKeepsAStock(t *testing.T) {
	gc := newTestBigBen()

	assert.Equal(t, CardCnt*2, BigBenTotalCards)
	for i, pile := range gc.GetFoundation() {
		require.Len(t, pile, 1, "clock %d starts with exactly its own card", i)
	}
	total := BigBenFoundationCnt
	for _, col := range gc.GetTableau() {
		assert.Len(t, col, BigBenColumnLen)
		total += len(col)
	}
	assert.Equal(t, BigBenFoundationCnt+BigBenTableauCnt*BigBenColumnLen, total)
	// 残りは山札。1 組ではこの枚数にならない。
	assert.Equal(t, CardCnt*2-total, gc.GetStockCount())
	assert.Equal(t, 52, gc.GetStockCount())
}

// **文字盤の並びと目標ランクが違う。**9 時が ♣2 で時計回りに 1 ランクずつ上がり、
// スートは ♣♥♠♦ の順。各文字盤は「自分の時刻のランクを表示したら完成」。
//
// **そしてその枚数がちょうど 104 に閉じる** ── これが偶然でないことは、
// 4 本が 8 枚・8 本が 9 枚で 4×8 + 8×9 = 104 になることで分かる。閉じない
// 並びを入れると、勝てないかカードが余るゲームになる。
func TestBigBen_ClockStartersCloseOnExactlyTheDeck(t *testing.T) {
	require.Len(t, bigBenStarters, BigBenFoundationCnt)

	total := 0
	for i, s := range bigBenStarters {
		target := BigBenTargetRank(i)
		// start から target まで、K→A で折り返しながら数える。
		n := 1
		for v := s.value; v != target; v = bigBenNextRank(v) {
			n++
			require.Less(t, n, CardValueMax+2, "clock %d never reaches its target", i)
		}
		total += n
	}
	assert.Equal(t, CardCnt*2, total, "文字盤の合計が 104 に閉じる")
}

// **タブローは同スート降順。**クローン元はスート不問なので、その述語を残すと
// 置けないはずの手が置ける。
func TestBigBen_TableauBuildsDownInSuit(t *testing.T) {
	gc := newTestBigBen()
	var fnd [BigBenFoundationCnt][]*Card
	// 降順なので、動かす札は置き先より 1 つ下。♠8 を ♠9 の上に置く。
	setBigBenBoard(gc, fnd, [][]*Card{
		{NewCard(CardDesignSpade, 8, true)},
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignHeart, 9, true)},
	})

	// 負のコントロール: ♥9 の上の ♠8 はクローン元（スート不問）なら合法。
	assert.Error(t, gc.MoveTableauToTableau(0, 2), "スートが違えば置けない")
	require.NoError(t, gc.MoveTableauToTableau(0, 1), "同スートなら置ける")
}

// **空列は埋められない。**クローン元は空列に何でも置ける。
func TestBigBen_AnEmptyColumnIsNotAPlace(t *testing.T) {
	gc := newTestBigBen()
	var fnd [BigBenFoundationCnt][]*Card
	setBigBenBoard(gc, fnd, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{},
	})

	assert.Error(t, gc.MoveTableauToTableau(0, 1), "空列には置けない")
}

// **手が尽きたら山札から補充する。**各列が 3 枚になるまで配り、すでに全列が
// 3 枚以上なら 1 巡だけ配る。クローン元に山札は無い。
func TestBigBen_DealTopsEveryColumnUpToThree(t *testing.T) {
	gc := newTestBigBen()
	var fnd [BigBenFoundationCnt][]*Card
	setBigBenBoard(gc, fnd, [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignSpade, 8, true), NewCard(CardDesignSpade, 7, true)},
	})
	before := gc.GetStockCount()

	require.NoError(t, gc.Deal())

	got := gc.GetTableau()
	assert.Len(t, got[0], BigBenDealMinimum, "1 枚の列は 3 枚まで積まれる")
	assert.Len(t, got[1], BigBenDealMinimum, "2 枚の列も 3 枚まで")
	// 空だった残り 6 列も 3 枚ずつ。合計 2+1+6*3 = 21 枚配られる。
	for i := 2; i < BigBenTableauCnt; i++ {
		assert.Len(t, got[i], BigBenDealMinimum, "col %d", i)
	}
	assert.Equal(t, before-21, gc.GetStockCount())
}

// **全列が 3 枚以上なら、1 巡だけ配る。**ここで何もしないと、山札が残って
// いるのに手が進まなくなる。
func TestBigBen_DealGivesOneRoundWhenEveryColumnIsAlreadyDeep(t *testing.T) {
	gc := newTestBigBen()
	var fnd [BigBenFoundationCnt][]*Card
	cols := make([][]*Card, BigBenTableauCnt)
	for i := range cols {
		cols[i] = []*Card{
			NewCard(CardDesignSpade, 9, true),
			NewCard(CardDesignHeart, 5, true),
			NewCard(CardDesignClover, 3, true),
		}
	}
	setBigBenBoard(gc, fnd, cols)
	before := gc.GetStockCount()

	require.NoError(t, gc.Deal())

	for i, col := range gc.GetTableau() {
		assert.Len(t, col, BigBenDealMinimum+1, "col %d gets exactly one more", i)
	}
	assert.Equal(t, before-BigBenTableauCnt, gc.GetStockCount())
}

// **山札が尽きたら配れない。**再配りは無い。
func TestBigBen_DealIsRefusedOnceTheStockIsEmpty(t *testing.T) {
	gc := newTestBigBen()
	gc.stock = nil

	err := gc.Deal()
	require.Error(t, err)
	code, _ := ErrorMessageCode(err)
	assert.Equal(t, "bigben.errStockEmptyNoRedeal", code)
}

// **補充も巻き戻せること。**スナップショットに山札を載せ忘れると、undo で札は
// 列から消えるのに山札へ戻らず、盤から枚数が失われる。クローン元は山札を
// 持たないので、この項そのものが無かった。
func TestBigBen_UndoingADealReturnsTheCardsToTheStock(t *testing.T) {
	gc := newTestBigBen()
	var fnd [BigBenFoundationCnt][]*Card
	setBigBenBoard(gc, fnd, [][]*Card{{NewCard(CardDesignSpade, 9, true)}})
	gc.stock = []*Card{
		NewCard(CardDesignHeart, 2, true),
		NewCard(CardDesignHeart, 3, true),
		NewCard(CardDesignHeart, 4, true),
	}
	countAll := func() int {
		n := gc.GetStockCount()
		for _, col := range gc.GetTableau() {
			n += len(col)
		}
		for _, pile := range gc.GetFoundation() {
			n += len(pile)
		}
		return n
	}
	before := countAll()

	require.NoError(t, gc.Deal())
	require.Equal(t, before, countAll(), "配っても総枚数は変わらない")

	require.NoError(t, gc.Undo())
	assert.Equal(t, 3, gc.GetStockCount(), "配った札は山札へ戻る")
	assert.Len(t, gc.GetTableau()[0], 1)
	assert.Equal(t, before, countAll(), "戻しても総枚数は変わらない")
}
