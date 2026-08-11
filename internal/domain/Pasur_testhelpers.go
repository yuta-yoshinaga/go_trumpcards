//go:build test

package domain

// SetCurrentPlayerIdxForTest は現在の手番を設定する
func (p *Pasur) SetCurrentPlayerIdxForTest(i int) { p.currentPlayerIdx = i }

// SetTableForTest は場の札を設定する
func (p *Pasur) SetTableForTest(cards []*Card) { p.tableCards = cards }

// SetLastCaptureIdxForTest は最後に捕獲した席を設定する
func (p *Pasur) SetLastCaptureIdxForTest(i int) { p.lastCaptureIdx = i }

// PlayForTest は指定プレイヤーに 1 枚出させる
func (p *Pasur) PlayForTest(playerIdx, cardIndex int, tableIndices []int) error {
	return p.play(playerIdx, cardIndex, tableIndices)
}

// CpuChoiceForTest は CPU が選ぶ手を返す
func (p *Pasur) CpuChoiceForTest(playerIdx int) (int, []int) { return p.chooseCpuMove(playerIdx) }

// FinishGameForTest は終局処理を直接呼ぶ
func (p *Pasur) FinishGameForTest() { p.finishGame() }

// ScoreOfForTest は 1 人の得点を返す
func (p *Pasur) ScoreOfForTest(i int) int { return p.scoreOf(p.players[i]) }

// DrainDeckForTest は山札を空にする
func (p *Pasur) DrainDeckForTest() {
	for p.trumpCards.GetRemainingCount() > 0 {
		p.trumpCards.DrawCard()
	}
}

// EmptyHandsForTest は全員の手札を空にする
func (p *Pasur) EmptyHandsForTest() {
	for _, pl := range p.players {
		pl.Reset()
	}
}

// pasurHandOf は playerIdx の手札を cards ちょうどに置き換える。
func pasurHandOf(p *Pasur, playerIdx int, cards ...*Card) {
	pl := p.GetPlayer(playerIdx)
	pl.Reset()
	for _, c := range cards {
		pl.AddCard(c)
	}
}
