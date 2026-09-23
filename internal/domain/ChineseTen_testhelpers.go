//go:build test

package domain

// SetLayoutForTest は場札を差し替える (テスト用)。
func (c *ChineseTen) SetLayoutForTest(cards []*Card) { c.layout = cards }

// SetCurrentPlayerForTest は手番を設定する (テスト用)。
func (c *ChineseTen) SetCurrentPlayerForTest(idx int) { c.currentIdx = idx }

// SetStockForTest は山札を差し替える (テスト用)。
func (c *ChineseTen) SetStockForTest(cards []*Card) { c.stock = cards }
