//go:build test

package domain

// ResolveShowdownForTest は手札を設定済みの状態でショーダウンを解決する (テスト用)。
func (g *Anaconda) ResolveShowdownForTest() { g.resolveShowdown() }
