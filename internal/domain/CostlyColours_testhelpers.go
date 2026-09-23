//go:build test

package domain

// SetTurnUpForTest は表の 1 枚を差し替える (テスト用)。
func (c *CostlyColours) SetTurnUpForTest(card *Card) { c.turnUp = card }

// SetTotalForTest は数え上げの累計を差し替える (テスト用)。
func (c *CostlyColours) SetTotalForTest(n int) { c.total = n }

// SetPhaseForTest はフェーズを差し替える (テスト用)。
func (c *CostlyColours) SetPhaseForTest(p string) { c.phase = p }

// SetCurrentForTest は手番の席を差し替える (テスト用)。
func (c *CostlyColours) SetCurrentForTest(i int) { c.currentIdx = i }

// FinishDealForTest はショーを数える (テスト用)。
func (c *CostlyColours) FinishDealForTest() { c.finishDeal() }
