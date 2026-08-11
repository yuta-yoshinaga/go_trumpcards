//go:build test

package domain

// SetPhaseForTest はフェーズを設定する
func (h *HoneymoonBridge) SetPhaseForTest(phase HoneymoonBridgePhase) { h.phase = phase }

// SetTrumpSuitForTest は切り札を設定する
func (h *HoneymoonBridge) SetTrumpSuitForTest(suit int) { h.trumpSuit = suit }

// SetDealerIdxForTest は親を設定する
func (h *HoneymoonBridge) SetDealerIdxForTest(i int) { h.dealerIdx = i }

// SetCurrentPlayerIdxForTest は現在の手番を設定する
func (h *HoneymoonBridge) SetCurrentPlayerIdxForTest(i int) { h.currentPlayerIdx = i }

// SetLeadPlayerIdxForTest はリードプレイヤーを設定する
func (h *HoneymoonBridge) SetLeadPlayerIdxForTest(i int) { h.leadPlayerIdx = i }

// SetCurrentTrickForTest は現在のトリックを設定する
func (h *HoneymoonBridge) SetCurrentTrickForTest(tc []*TrickCard) { h.currentTrick = tc }

// SetContractForTest は落札者と契約を設定する
func (h *HoneymoonBridge) SetContractForTest(declarer, level, suit int) {
	h.declarerIdx, h.contractLevel, h.trumpSuit = declarer, level, suit
}

// SetTrickNumberForTest はトリック数を設定する
func (h *HoneymoonBridge) SetTrickNumberForTest(n int) { h.trickNumber = n }

// SetStockForTest は山札を設定する
func (h *HoneymoonBridge) SetStockForTest(cards []*Card) { h.stock = cards }

// GiveTricksForTest は指定プレイヤーに空のトリックを n 個持たせる
func (h *HoneymoonBridge) GiveTricksForTest(playerIdx, n int) {
	for range n {
		h.players[playerIdx].AddTrick([]*Card{})
	}
}

// PlayForTest は指定プレイヤーに 1 枚出させる
func (h *HoneymoonBridge) PlayForTest(playerIdx, cardIndex int) error {
	return h.play(playerIdx, cardIndex)
}

// BidForTest は指定プレイヤーに宣言させる
func (h *HoneymoonBridge) BidForTest(playerIdx, level, suit int) error {
	return h.bidBy(playerIdx, level, suit)
}

// CpuChoiceForTest は CPU が選ぶ手札インデックスを返す
func (h *HoneymoonBridge) CpuChoiceForTest(playerIdx int) int { return h.chooseCpuCard(playerIdx) }

// CpuBidChoiceForTest は CPU が選ぶ契約を返す
func (h *HoneymoonBridge) CpuBidChoiceForTest(playerIdx int) (int, int) {
	return h.chooseCpuBid(playerIdx)
}

// FinishRoundForTest はディール精算を直接呼ぶ
func (h *HoneymoonBridge) FinishRoundForTest() { h.finishRound() }

// FinishGameForTest は終局処理を直接呼ぶ
func (h *HoneymoonBridge) FinishGameForTest() { h.finishGame() }

// TrickWinnerForTest はトリックの勝者を返す
func (h *HoneymoonBridge) TrickWinnerForTest() int { return h.trickWinner() }

// OutbidsForTest は宣言が現在の契約を上回るかを返す
func (h *HoneymoonBridge) OutbidsForTest(level, suit int) bool { return h.outbids(level, suit) }

// honeymoonBridgeHandOf は playerIdx の手札を cards ちょうどに置き換える。
func honeymoonBridgeHandOf(h *HoneymoonBridge, playerIdx int, cards ...*Card) {
	p := h.GetPlayer(playerIdx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}
