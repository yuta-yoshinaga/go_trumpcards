//go:build test

package domain

// SetTrumpSuitForTest は切札を設定する (テスト用)。
func (s *Skitgubbe) SetTrumpSuitForTest(suit int) { s.trumpSuit = suit }

// SetPhaseForTest はフェーズを設定する (テスト用)。
func (s *Skitgubbe) SetPhaseForTest(p SkitgubbePhase) { s.phase = p }

// SetPileForTest は第2フェーズの場札を設定する (テスト用)。
func (s *Skitgubbe) SetPileForTest(cards []*Card) { s.pile = cards }

// SetStockForTest は山札を設定する (テスト用)。
func (s *Skitgubbe) SetStockForTest(cards []*Card) { s.stock = cards }

// SetCurrentPlayerForTest は手番を設定する (テスト用)。
func (s *Skitgubbe) SetCurrentPlayerForTest(idx int) { s.currentIdx = idx }
