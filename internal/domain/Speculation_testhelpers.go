//go:build test

package domain

// DeckForTest は山札を返す（テスト用）。
func (g *Speculation) DeckForTest() *TrumpCards { return g.deck }
