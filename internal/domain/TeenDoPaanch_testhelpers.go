//go:build test

package domain

// SetPhaseForTest はフェーズを設定する
func (g *TeenDoPaanch) SetPhaseForTest(phase TeenDoPaanchPhase) { g.phase = phase }

// SetTrumpSuitForTest は切り札を設定する
func (g *TeenDoPaanch) SetTrumpSuitForTest(suit int) { g.trumpSuit = suit }

// SetFivePlayerIdxForTest はノルマ 5 の席を設定する
func (g *TeenDoPaanch) SetFivePlayerIdxForTest(i int) { g.fivePlayerIdx = i }

// SetCurrentPlayerIdxForTest は現在の手番を設定する
func (g *TeenDoPaanch) SetCurrentPlayerIdxForTest(i int) { g.currentPlayerIdx = i }

// SetLeadPlayerIdxForTest はリードプレイヤーを設定する
func (g *TeenDoPaanch) SetLeadPlayerIdxForTest(i int) { g.leadPlayerIdx = i }

// SetCurrentTrickForTest は現在のトリックを設定する
func (g *TeenDoPaanch) SetCurrentTrickForTest(tc []*TrickCard) { g.currentTrick = tc }

// SetSurplusForTest は過不足を設定する
func (g *TeenDoPaanch) SetSurplusForTest(s []int) { g.surplus = s }

// GetSurplusForTest は過不足を返す
func (g *TeenDoPaanch) GetSurplusForTest() []int { return g.surplus }

// GiveTricksForTest は指定プレイヤーに空のトリックを n 個持たせる
func (g *TeenDoPaanch) GiveTricksForTest(playerIdx, n int) {
	for range n {
		g.players[playerIdx].AddTrick([]*Card{})
	}
}

// PlayForTest は指定プレイヤーに 1 枚出させる
func (g *TeenDoPaanch) PlayForTest(playerIdx, cardIndex int) error {
	return g.play(playerIdx, cardIndex)
}

// CpuChoiceForTest は CPU が選ぶ手札インデックスを返す
func (g *TeenDoPaanch) CpuChoiceForTest(playerIdx int) int { return g.chooseCpuCard(playerIdx) }

// CpuTrumpChoiceForTest は CPU が選ぶ切り札を返す
func (g *TeenDoPaanch) CpuTrumpChoiceForTest(playerIdx int) int { return g.chooseCpuTrump(playerIdx) }

// FinishRoundForTest はラウンド精算を直接呼ぶ
func (g *TeenDoPaanch) FinishRoundForTest() { g.finishRound() }

// FinishGameForTest は終局処理を直接呼ぶ
func (g *TeenDoPaanch) FinishGameForTest() { g.finishGame() }

// TrickWinnerForTest はトリックの勝者を返す
func (g *TeenDoPaanch) TrickWinnerForTest() int { return g.trickWinner() }

// teenDoPaanchHandOf は playerIdx の手札を cards ちょうどに置き換える。
//
// **配りの上に積んではいけない。** 残った札が混ざると配り依存で落ちる。
func teenDoPaanchHandOf(g *TeenDoPaanch, playerIdx int, cards ...*Card) {
	p := g.GetPlayer(playerIdx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// ExchangeForTest は札のやり取りを直接呼ぶ
func (g *TeenDoPaanch) ExchangeForTest() { g.exchangeCards() }
