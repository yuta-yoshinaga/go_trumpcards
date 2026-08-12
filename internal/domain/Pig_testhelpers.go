//go:build test

package domain

// SetPhaseForTest はフェーズを設定する
func (g *Pig) SetPhaseForTest(p PigPhase) { g.phase = p }

// SetCurrentPlayerIdxForTest は渡す番の席を設定する
func (g *Pig) SetCurrentPlayerIdxForTest(i int) { g.currentPlayerIdx = i }

// GiveHandForTest は指定席の手札を cards ちょうどに置き換える
func (g *Pig) GiveHandForTest(playerIdx int, cards ...*Card) {
	p := g.players[playerIdx]
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// ChoosePassForTest は指定席に渡す札を選ばせる
func (g *Pig) ChoosePassForTest(playerIdx, cardIndex int) error {
	return g.choosePass(playerIdx, cardIndex)
}

// CpuChoiceForTest は CPU が渡す手札インデックスを返す
func (g *Pig) CpuChoiceForTest(playerIdx int) int { return g.chooseCpuCard(playerIdx) }

// NoticeForTest は指定席を気づいた扱いにする
func (g *Pig) NoticeForTest(playerIdx int) { g.notice(playerIdx) }

// OpenSignalForTest は合図フェーズを開く
func (g *Pig) OpenSignalForTest(playerIdx int) { g.openSignal(playerIdx) }

// PigRanksForTest は n 人卓で使うランクを返す
func PigRanksForTest(n int) []int { return pigRanksFor(n) }

// HumanSeatForTest は人間の席を返す
func (g *Pig) HumanSeatForTest() int { return g.humanSeat() }

// SetRoundLoserIdxForTest は直近ラウンドの敗者を設定する
func (g *Pig) SetRoundLoserIdxForTest(i int) { g.roundLoserIdx = i }
