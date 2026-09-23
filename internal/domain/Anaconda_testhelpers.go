//go:build test

package domain

// ResolveShowdownForTest は手札を設定済みの状態でショーダウンを解決する (テスト用)。
func (g *Anaconda) ResolveShowdownForTest() { g.resolveShowdown() }

// EnterRollForTest はロールフェーズ入場 (最初のベッティングラウンド開始 + CPU 進行) を
// 手動で発火する (テスト用)。
func (g *Anaconda) EnterRollForTest() { g.enterRollPhase() }
