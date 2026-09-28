//go:build test

package domain

// SetPlayStateForTest はプレイ状態 (シーケンス・手番) を直接設定する (テスト用)。
// 直前のラウンド解決状態 (scored / winnerIdx / result) もクリアし、決定的なプレイ
// シナリオを隔離して組み立てられるようにする。
func (g *Michigan) SetPlayStateForTest(seqSuit, seqHigh, current, last int) {
	g.state.phase = MichiganPhasePlay
	g.state.seqSuit = seqSuit
	g.state.seqHighValue = seqHigh
	g.state.currentPlayer = current
	g.state.lastPlayerIdx = last
	g.state.scored = false
	g.state.winnerIdx = -1
	g.state.result = MichiganResultNone
}

// AddDeadCardForTest はデッドハンドにカードを追加する (テスト用)。
func (g *Michigan) AddDeadCardForTest(c *Card) {
	g.state.deadHand = append(g.state.deadHand, c)
}

// SetBoodleForTest はブードルのチップ・獲得状態を設定する (テスト用)。
func (g *Michigan) SetBoodleForTest(i, chips, claimedBy int) {
	if i < 0 || i >= len(g.state.boodles) {
		return
	}
	g.state.boodles[i].chips = chips
	g.state.boodles[i].claimedBy = claimedBy
}

// SetRoundStartChipsForTest はラウンド開始チップを設定する (テスト用)。
func (g *Michigan) SetRoundStartChipsForTest(chips []int) {
	g.state.roundStartChips = chips
}

// DoPlayForTest は seat の手札インデックス idx を出す (乱数配札を迂回した決定的検証用)。
func (g *Michigan) DoPlayForTest(seat, idx int) { g.doPlay(seat, idx) }
