//go:build test

package domain

// SetPhaseForTest はフェーズを設定する
func (s *SergeantMajor) SetPhaseForTest(phase SergeantMajorPhase) { s.phase = phase }

// SetTrumpSuitForTest は切り札を設定する
func (s *SergeantMajor) SetTrumpSuitForTest(suit int) { s.trumpSuit = suit }

// SetDealerIdxForTest は親を設定する
func (s *SergeantMajor) SetDealerIdxForTest(i int) { s.dealerIdx = i }

// SetCurrentPlayerIdxForTest は現在の手番を設定する
func (s *SergeantMajor) SetCurrentPlayerIdxForTest(i int) { s.currentPlayerIdx = i }

// SetLeadPlayerIdxForTest はリードプレイヤーを設定する
func (s *SergeantMajor) SetLeadPlayerIdxForTest(i int) { s.leadPlayerIdx = i }

// SetCurrentTrickForTest は現在のトリックを設定する
func (s *SergeantMajor) SetCurrentTrickForTest(tc []*TrickCard) { s.currentTrick = tc }

// SetSurplusForTest は過不足を設定する
func (s *SergeantMajor) SetSurplusForTest(v []int) { s.surplus = v }

// GetSurplusForTest は過不足を返す
func (s *SergeantMajor) GetSurplusForTest() []int { return s.surplus }

// GiveTricksForTest は指定プレイヤーに空のトリックを n 個持たせる
func (s *SergeantMajor) GiveTricksForTest(playerIdx, n int) {
	for range n {
		s.players[playerIdx].AddTrick([]*Card{})
	}
}

// PlayForTest は指定プレイヤーに 1 枚出させる
func (s *SergeantMajor) PlayForTest(playerIdx, cardIndex int) error {
	return s.play(playerIdx, cardIndex)
}

// DiscardForTest は親に指定の札を捨てさせる
func (s *SergeantMajor) DiscardForTest(playerIdx int, indices []int) error {
	return s.discardBy(playerIdx, indices)
}

// CpuChoiceForTest は CPU が選ぶ手札インデックスを返す
func (s *SergeantMajor) CpuChoiceForTest(playerIdx int) int { return s.chooseCpuCard(playerIdx) }

// CpuTrumpChoiceForTest は CPU が選ぶ切り札を返す
func (s *SergeantMajor) CpuTrumpChoiceForTest(playerIdx int) int { return s.chooseCpuTrump(playerIdx) }

// CpuDiscardChoiceForTest は CPU が捨てる札を返す
func (s *SergeantMajor) CpuDiscardChoiceForTest(playerIdx int) []int {
	return s.chooseCpuDiscard(playerIdx)
}

// ExchangeForTest は札のやり取りを直接呼ぶ
func (s *SergeantMajor) ExchangeForTest() { s.exchangeCards() }

// FinishRoundForTest はラウンド精算を直接呼ぶ
func (s *SergeantMajor) FinishRoundForTest() { s.finishRound() }

// FinishGameForTest は終局処理を直接呼ぶ
func (s *SergeantMajor) FinishGameForTest() { s.finishGame() }

// TrickWinnerForTest はトリックの勝者を返す
func (s *SergeantMajor) TrickWinnerForTest() int { return s.trickWinner() }

// sergeantMajorHandOf は playerIdx の手札を cards ちょうどに置き換える。
func sergeantMajorHandOf(s *SergeantMajor, playerIdx int, cards ...*Card) {
	p := s.GetPlayer(playerIdx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}
