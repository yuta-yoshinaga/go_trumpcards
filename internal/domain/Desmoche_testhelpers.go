//go:build test

package domain

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (d *Desmoche) SetPhaseForTest(p DesmochePhase) { d.phase = p }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (d *Desmoche) SetCurrentPlayerForTest(idx int) { d.currentIdx = idx }

// SetStockForTest はテスト用に山札を差し替える。
func (d *Desmoche) SetStockForTest(cards []*Card) { d.stock = cards }

// SetDiscardForTest はテスト用に捨て札を差し替える。
func (d *Desmoche) SetDiscardForTest(cards []*Card) { d.discard = cards }

// SetRoundNumberForTest はテスト用にラウンド数を差し替える。
func (d *Desmoche) SetRoundNumberForTest(n int) { d.roundNo = n }
