//go:build test

package domain

// SetChipsForTest はテスト用に残高を差し替える。
func (p *BaccaratBanquePlayer) SetChipsForTest(n int) { p.chips = n }
