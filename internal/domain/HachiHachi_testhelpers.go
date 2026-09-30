//go:build test

package domain

// SetActionLogForTest は直近イベント表示を検証するため棋譜を設定する。
func (g *HachiHachi) SetActionLogForTest(entries []*ActionLogEntry) { g.state.actionLog = entries }

// HachiHachiCardPoints は札の素点を返す (試験用)。
func HachiHachiCardPoints(c *Card) int { return hachihachiCardPoints(c) }
