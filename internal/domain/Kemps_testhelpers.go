//go:build test

package domain

// SetGameEndFlagForTest はテスト用に終了フラグを設定する。
func (g *Kemps) SetGameEndFlagForTest(v bool) { g.gameEndFlag = v }
