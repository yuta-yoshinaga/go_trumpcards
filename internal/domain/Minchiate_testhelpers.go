//go:build test

package domain

// SetRoundEndForTest は精算内訳の表示を確かめるための盤を組む (テスト用)。
//
// 最終トリックの勝者とスカルト枚数は**その局の進み方でしか決まらない**ので、
// 表示だけを見たい場面では直接置くほうが配りに依存しない。
func (g *Minchiate) SetRoundEndForTest(lastTrickWinner, dealerIdx, scartoSize int) {
	g.lastTrickWinner = lastTrickWinner
	g.dealerIdx = dealerIdx
	g.scarto = make([]*Card, scartoSize)
}
