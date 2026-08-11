//go:build test

package domain

// SetPhaseForTest はフェーズを設定する
func (m *Minibridge) SetPhaseForTest(phase MinibridgePhase) { m.phase = phase }

// SetDealerIdxForTest は親を設定する
func (m *Minibridge) SetDealerIdxForTest(i int) { m.dealerIdx = i }

// SetCurrentPlayerIdxForTest は現在の手番を設定する
func (m *Minibridge) SetCurrentPlayerIdxForTest(i int) { m.currentPlayerIdx = i }

// SetLeadPlayerIdxForTest はリードプレイヤーを設定する
func (m *Minibridge) SetLeadPlayerIdxForTest(i int) { m.leadPlayerIdx = i }

// SetCurrentTrickForTest は現在のトリックを設定する
func (m *Minibridge) SetCurrentTrickForTest(tc []*TrickCard) { m.currentTrick = tc }

// SetContractForTest は落札者と契約を設定する
func (m *Minibridge) SetContractForTest(declarer, level, suit int) {
	m.declarerIdx, m.contractLevel, m.contractSuit = declarer, level, suit
	m.dummyIdx = (declarer + MinibridgeTeamCnt) % MinibridgePlayerCnt
}

// SetRoundNumberForTest はディール番号を設定する
func (m *Minibridge) SetRoundNumberForTest(n int) { m.roundNumber = n }

// GiveTricksForTest は指定プレイヤーに空のトリックを n 個持たせる
func (m *Minibridge) GiveTricksForTest(playerIdx, n int) {
	for range n {
		m.players[playerIdx].AddTrick([]*Card{})
	}
}

// PlayForTest は指定プレイヤーに 1 枚出させる
func (m *Minibridge) PlayForTest(playerIdx, cardIndex int) error {
	return m.play(playerIdx, cardIndex)
}

// SelectContractForTest は落札者に契約を選ばせる
func (m *Minibridge) SelectContractForTest(playerIdx, level, suit int) error {
	return m.selectContractBy(playerIdx, level, suit)
}

// CpuChoiceForTest は CPU が選ぶ手札インデックスを返す
func (m *Minibridge) CpuChoiceForTest(playerIdx int) int { return m.chooseCpuCard(playerIdx) }

// CpuContractChoiceForTest は CPU が選ぶ契約を返す
func (m *Minibridge) CpuContractChoiceForTest(playerIdx int) (int, int) {
	return m.chooseCpuContract(playerIdx)
}

// FinishRoundForTest はディール精算を直接呼ぶ
func (m *Minibridge) FinishRoundForTest() { m.finishRound() }

// FinishGameForTest は終局処理を直接呼ぶ
func (m *Minibridge) FinishGameForTest() { m.finishGame() }

// TrickWinnerForTest はトリックの勝者を返す
func (m *Minibridge) TrickWinnerForTest() int { return m.trickWinner() }

// ContractPointsForTest は契約点（ボーナス抜き）を返す
func (m *Minibridge) ContractPointsForTest() int { return m.contractPoints() }

// DecideDeclarerForTest は落札者決定を直接呼ぶ
func (m *Minibridge) DecideDeclarerForTest() { m.decideDeclarer() }

// SetHcpForTest は各席の HCP を直接設定する（合計 40 になるかは呼び手の責任）
func (m *Minibridge) SetHcpForTest(hcp [MinibridgePlayerCnt]int) {
	for i, v := range hcp {
		m.players[i].SetHcp(v)
	}
}

// minibridgeHandOf は playerIdx の手札を cards ちょうどに置き換える。
func minibridgeHandOf(m *Minibridge, playerIdx int, cards ...*Card) {
	p := m.GetPlayer(playerIdx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}
