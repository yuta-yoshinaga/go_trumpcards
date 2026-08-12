//go:build test

package domain

// SetCurrentPlayerIdxForTest は現在の手番を設定する
func (r *RollingStone) SetCurrentPlayerIdxForTest(i int) { r.currentPlayerIdx = i }

// SetLeadPlayerIdxForTest はリード席を設定する
func (r *RollingStone) SetLeadPlayerIdxForTest(i int) { r.leadPlayerIdx = i }

// SetCurrentTrickForTest は現在のトリックを設定する
func (r *RollingStone) SetCurrentTrickForTest(tc []*TrickCard) { r.currentTrick = tc }

// PlayForTest は指定プレイヤーに 1 枚出させる
func (r *RollingStone) PlayForTest(playerIdx, cardIndex int) error {
	return r.play(playerIdx, cardIndex)
}

// PickUpForTest は指定プレイヤーに引き取らせる
func (r *RollingStone) PickUpForTest(playerIdx int) error { return r.pickUp(playerIdx) }

// CpuChoiceForTest は CPU が選ぶ手札インデックスを返す
func (r *RollingStone) CpuChoiceForTest(playerIdx int) int { return r.chooseCpuCard(playerIdx) }

// TrickWinnerForTest はトリックの勝者を返す
func (r *RollingStone) TrickWinnerForTest() int { return r.trickWinner() }

// GiveHandForTest は指定席の手札を cards ちょうどに置き換える
func (r *RollingStone) GiveHandForTest(playerIdx int, cards ...*Card) {
	p := r.players[playerIdx]
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}
