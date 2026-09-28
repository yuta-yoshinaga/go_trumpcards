//go:build test

package domain

// StartPlayForTest はテスト用にラフとメルドを精算してプレイフェーズを開始する。
func (g *Gleek) StartPlayForTest() { g.startPlay() }

// SetTrickPointsForTest はテスト用にトリック点を差し込む。
func (g *Gleek) SetTrickPointsForTest(p [GleekPlayerCnt]int) { g.trickPoints = p }

// ScoreMeldsForTest はテスト用にメルドだけを精算する。
func (g *Gleek) ScoreMeldsForTest() { g.scoreMelds() }
