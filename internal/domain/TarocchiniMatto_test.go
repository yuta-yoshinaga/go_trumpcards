//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// マットはどのトリックも取らない。**リードされた場合も同じ。**
// ここは ledSuit() を経由する本番の経路で確かめる —— 直接 led を渡すと、
// リードスートの決め方の誤りをテストが迂回してしまう。
func TestTarocchini_MattoLedNeverTakesTheTrick(t *testing.T) {
	g := NewDefaultTarocchini()
	g.Reset()
	g.SetPhase(TarocchiniPhaseTrickEnd)
	g.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(TarocchiniMattoDesign, TarocchiniMattoValue, false)},
		{PlayerIdx: 1, Card: NewCard(1, 6, false)},
		{PlayerIdx: 2, Card: NewCard(1, 9, false)},
		{PlayerIdx: 3, Card: NewCard(2, 14, false)},
	}
	g.ResolveTrick()
	assert.Equal(t, 2, g.GetLastTrickWinner(),
		"the Matto must not take the trick; the highest card of the suit actually led should")
}

// マットがリードされたとき、スートはその次に出た札が決める。
func TestTarocchini_MattoLedDoesNotSetTheSuit(t *testing.T) {
	g := NewDefaultTarocchini()
	g.Reset()
	g.SetPhase(TarocchiniPhasePlay)
	// 席 0 が人間。マットをリードするのは席 3 にする。
	g.SetCurrentPlayerIdx(0)
	p := g.GetPlayer(0)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	p.AddCard(NewCard(1, 6, false))
	p.AddCard(NewCard(TarocchiniTrumpDesign, 9, false))
	g.currentTrick = []*TrickCard{
		{PlayerIdx: 3, Card: NewCard(TarocchiniMattoDesign, TarocchiniMattoValue, false)},
	}
	// **切札を出す義務は生じない。**マットはスートを定めないので、次に出す人が
	// 自由にリードできる。切札だけが合法になると、手札を丸ごと縛ることになる。
	assert.Equal(t, []int{0, 1}, g.GetValidPlayIndices(0))
	require.NoError(t, g.PlayerPlay(0), "a plain card must be legal after a led Matto")
}

// 味方がマットをリードした局面は「味方が勝っている」ではない。マットは
// トリックを取らないので、まだ誰のものでもない。ここを follow 扱いにすると
// 取りに行くべきトリックをダックしてしまう。
func TestTarocchini_ALoneLedMattoLeavesTheTrickOpen(t *testing.T) {
	g := NewDefaultTarocchini()
	g.Reset()
	g.SetPhase(TarocchiniPhasePlay)
	g.SetCurrentPlayerIdx(0)
	p := g.GetPlayer(0)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	p.AddCard(NewCard(1, 6, false))                      // weak
	p.AddCard(NewCard(TarocchiniTrumpDesign, 20, false)) // strong
	// 席 2 は席 0 のパートナー (対面)。
	require.Equal(t, TarocchiniTeamOf(0), TarocchiniTeamOf(2))
	g.currentTrick = []*TrickCard{
		{PlayerIdx: 2, Card: NewCard(TarocchiniMattoDesign, TarocchiniMattoValue, false)},
	}

	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, []int{1}, hint.CardIndices,
		"the trick is still open, so the strong card should be suggested rather than a duck")
}
