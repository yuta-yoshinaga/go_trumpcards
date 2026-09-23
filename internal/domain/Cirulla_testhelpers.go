//go:build test

package domain

// SetTableForTest は場札を差し替える (テスト用)。
//
// **場は配りで決まるので、狙った盤面は組めない。** 捕獲規則を確かめるには
// ここで固定するしかない。
func (c *Cirulla) SetTableForTest(cards []*Card) { c.table = cards }
