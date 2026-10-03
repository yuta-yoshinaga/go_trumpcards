//go:build test

package domain

// IndianRummyCardPoints はデッドウッド計算に使うカード点を返す (試験用)。
func IndianRummyCardPoints(card *Card, wildRank int) int {
	return indianRummyCardPoints(card, wildRank)
}
