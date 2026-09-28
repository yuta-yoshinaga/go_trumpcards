//go:build test

package domain

// SetShoeForTest はシューを差し替える (テスト用)。
func (b *BaccaratBanque) SetShoeForTest(cards []*Card) {
	b.shoe = cards
	b.drawIdx = 0
}

// SetPhaseForTest はフェーズを差し替える (テスト用)。
func (b *BaccaratBanque) SetPhaseForTest(p string) { b.phase = p }

// SettleForTest は決着させる (テスト用)。
func (b *BaccaratBanque) SettleForTest() { b.settle() }
