//go:build test

package domain

// SetCurrentPlayerIdxForTest は現在の手番を設定する
func (l *LingerLonger) SetCurrentPlayerIdxForTest(i int) { l.currentPlayerIdx = i }

// SetLeadPlayerIdxForTest はリード席を設定する
func (l *LingerLonger) SetLeadPlayerIdxForTest(i int) { l.leadPlayerIdx = i }

// SetCurrentTrickForTest は現在のトリックを設定する
func (l *LingerLonger) SetCurrentTrickForTest(tc []*TrickCard) { l.currentTrick = tc }

// PlayForTest は指定プレイヤーに 1 枚出させる
func (l *LingerLonger) PlayForTest(playerIdx, cardIndex int) error {
	return l.play(playerIdx, cardIndex)
}

// CpuChoiceForTest は CPU が選ぶ手札インデックスを返す
func (l *LingerLonger) CpuChoiceForTest(playerIdx int) int { return l.chooseCpuCard(playerIdx) }

// TrickWinnerForTest はトリックの勝者を返す
func (l *LingerLonger) TrickWinnerForTest() int { return l.trickWinner() }

// GiveHandForTest は指定席の手札を cards ちょうどに置き換える
func (l *LingerLonger) GiveHandForTest(playerIdx int, cards ...*Card) {
	p := l.players[playerIdx]
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// DrainStockForTest は山札を空にする
func (l *LingerLonger) DrainStockForTest() {
	for l.trumpCards.GetRemainingCount() > 0 {
		l.trumpCards.DrawCard()
	}
}
