//go:build test

package domain

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (l *Loba) SetPhaseForTest(p LobaPhase) { l.phase = p }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (l *Loba) SetCurrentPlayerForTest(idx int) { l.currentIdx = idx }

// SetStockForTest はテスト用に山札を差し替える。
func (l *Loba) SetStockForTest(cards []*Card) { l.stock = cards }

// SetDiscardForTest はテスト用に捨て札を差し替える。
func (l *Loba) SetDiscardForTest(cards []*Card) { l.discard = cards }

// SetScoreForTest はテスト用に失点を差し替える。
func (l *Loba) SetScoreForTest(idx, score int) { l.scores[idx] = score }

// SetHasMeldedForTest はテスト用にメルド済みフラグを差し替える。
func (l *Loba) SetHasMeldedForTest(idx int, v bool) { l.hasMelded[idx] = v }
