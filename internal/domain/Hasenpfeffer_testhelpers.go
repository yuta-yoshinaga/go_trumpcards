//go:build test

package domain

// SetPhaseForTest はフェーズを設定する
func (h *Hasenpfeffer) SetPhaseForTest(phase HasenpfefferPhase) { h.phase = phase }

// SetTrumpSuitForTest は切り札を設定する
func (h *Hasenpfeffer) SetTrumpSuitForTest(suit int) { h.trumpSuit = suit }

// SetDealerIdxForTest はディーラーを設定する
func (h *Hasenpfeffer) SetDealerIdxForTest(i int) { h.dealerIdx = i }

// SetCurrentPlayerIdxForTest は現在の手番を設定する
func (h *Hasenpfeffer) SetCurrentPlayerIdxForTest(i int) { h.currentPlayerIdx = i }

// SetLeadPlayerIdxForTest はリードプレイヤーを設定する
func (h *Hasenpfeffer) SetLeadPlayerIdxForTest(i int) { h.leadPlayerIdx = i }

// SetCurrentTrickForTest は現在のトリックを設定する
func (h *Hasenpfeffer) SetCurrentTrickForTest(tc []*TrickCard) { h.currentTrick = tc }

// SetContractForTest は落札者と落札額を設定する
func (h *Hasenpfeffer) SetContractForTest(declarer, contract int) {
	h.declarerIdx, h.contract = declarer, contract
}

// SetTrickNumberForTest はトリック数を設定する
func (h *Hasenpfeffer) SetTrickNumberForTest(n int) { h.trickNumber = n }

// GiveTricksForTest は指定プレイヤーに空のトリックを n 個持たせる
func (h *Hasenpfeffer) GiveTricksForTest(playerIdx, n int) {
	for range n {
		h.players[playerIdx].AddTrick([]*Card{})
	}
}

// BidForTest は指定プレイヤーに宣言させる
func (h *Hasenpfeffer) BidForTest(playerIdx, bid int) error { return h.bidBy(playerIdx, bid) }

// DiscardForTest は落札者に切り札を宣言させ 1 枚捨てさせる
func (h *Hasenpfeffer) DiscardForTest(playerIdx, cardIndex, suit int) error {
	return h.discardBy(playerIdx, cardIndex, suit)
}

// PlayForTest は指定プレイヤーに 1 枚出させる
func (h *Hasenpfeffer) PlayForTest(playerIdx, cardIndex int) error {
	return h.play(playerIdx, cardIndex)
}

// CpuChoiceForTest は CPU が選ぶ手札インデックスを返す
func (h *Hasenpfeffer) CpuChoiceForTest(playerIdx int) int { return h.chooseCpuCard(playerIdx) }

// CpuBidChoiceForTest は CPU が選ぶ宣言額を返す
func (h *Hasenpfeffer) CpuBidChoiceForTest(playerIdx int) int { return h.chooseCpuBid(playerIdx) }

// CpuTrumpChoiceForTest は CPU が選ぶ切り札を返す
func (h *Hasenpfeffer) CpuTrumpChoiceForTest(playerIdx int) int { return h.chooseCpuTrump(playerIdx) }

// FinishHandForTest はハンド精算を直接呼ぶ
func (h *Hasenpfeffer) FinishHandForTest() { h.finishHand() }

// FinishGameForTest は終局処理を直接呼ぶ
func (h *Hasenpfeffer) FinishGameForTest() { h.finishGame() }

// TrickWinnerForTest はトリックの勝者を返す
func (h *Hasenpfeffer) TrickWinnerForTest() int { return h.trickWinner() }

// hasenpfefferHandOf は playerIdx の手札を cards ちょうどに置き換える。
func hasenpfefferHandOf(h *Hasenpfeffer, playerIdx int, cards ...*Card) {
	p := h.GetPlayer(playerIdx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// CpuDiscardChoiceForTest は CPU が捨てる札のインデックスを返す
func (h *Hasenpfeffer) CpuDiscardChoiceForTest(playerIdx, suit int) int {
	return h.chooseCpuDiscard(playerIdx, suit)
}
