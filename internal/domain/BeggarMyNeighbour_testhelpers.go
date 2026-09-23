//go:build test

package domain

// SetRoundsPlayedForTest はテスト用に消化ラウンド数を設定する。
func (g *BeggarMyNeighbour) SetRoundsPlayedForTest(n int) { g.roundsPlayed = n }
