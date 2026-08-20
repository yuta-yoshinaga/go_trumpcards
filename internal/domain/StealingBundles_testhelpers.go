//go:build test

package domain

// SetCurrentPlayerIdxForTest は現在の手番を設定する
func (s *StealingBundles) SetCurrentPlayerIdxForTest(i int) { s.currentPlayerIdx = i }

// SetTableCardsForTest は場札を設定する
func (s *StealingBundles) SetTableCardsForTest(cards []*Card) { s.tableCards = cards }

// SetLastCaptureVictimIdxForTest は直前の盗みの被害者席を設定する (改竄の再現用)
func (s *StealingBundles) SetLastCaptureVictimIdxForTest(i int) { s.lastCaptureVictimIdx = i }

// GiveHandForTest は指定席の手札を cards ちょうどに置き換える
func (s *StealingBundles) GiveHandForTest(playerIdx int, cards ...*Card) {
	p := s.players[playerIdx]
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// TakeForTest は指定席に場札を取らせる
func (s *StealingBundles) TakeForTest(playerIdx, cardIndex int) error {
	return s.take(playerIdx, cardIndex)
}

// StealForTest は指定席に束を奪わせる
func (s *StealingBundles) StealForTest(playerIdx, cardIndex, victimIdx int) error {
	return s.steal(playerIdx, cardIndex, victimIdx)
}

// TrailForTest は指定席に場へ置かせる
func (s *StealingBundles) TrailForTest(playerIdx, cardIndex int) error {
	return s.trail(playerIdx, cardIndex)
}

// CpuChoiceForTest は CPU が選ぶ手を返す (手札の位置, 奪う相手 / -1)
func (s *StealingBundles) CpuChoiceForTest(playerIdx int) (int, int) {
	return s.chooseCpuMove(playerIdx)
}

// DrainDeckForTest は山札を空にする
func (s *StealingBundles) DrainDeckForTest() {
	for s.trumpCards.GetRemainingCount() > 0 {
		s.trumpCards.DrawCard()
	}
}
