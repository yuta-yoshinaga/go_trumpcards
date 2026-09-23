//go:build test

package domain

// SetStockForTest はテスト用に山札を差し替える。
func (c *ContinentalRummy) SetStockForTest(cards []*Card) { c.stock = cards }

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (c *ContinentalRummy) SetPhaseForTest(p string) { c.phase = p }

// SetCurrentIdxForTest はテスト用に手番を差し替える。
func (c *ContinentalRummy) SetCurrentIdxForTest(i int) { c.currentIdx = i }

// discardCountForTest はテスト用に捨て札の枚数を返す。
func (c *ContinentalRummy) discardCountForTest() int { return len(c.discardPile) }

// SetDiscardForTest はテスト用に捨て札を差し替える。
func (c *ContinentalRummy) SetDiscardForTest(cards []*Card) { c.discardPile = cards }
