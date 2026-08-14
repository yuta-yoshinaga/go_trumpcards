//go:build test

package domain

// SetCenterPileForTest は場札を設定する
func (g *Snap) SetCenterPileForTest(cards []*Card) { g.centerPile = cards }

// SetCurrentTurnIdxForTest はめくる席を設定する
func (g *Snap) SetCurrentTurnIdxForTest(i int) { g.currentTurnIdx = i }

// SetPendingForTest は保留アクションを設定する
func (g *Snap) SetPendingForTest(p SnapPending) { g.pending = p }

// StepForTest は指定席に 1 枚めくらせる
func (g *Snap) StepForTest(playerIdx int) { g.step(playerIdx) }

// ScheduleNextForTest は次の予約を直接呼ぶ
func (g *Snap) ScheduleNextForTest() { g.scheduleNext() }

// ReactionMsForTest は反応時間を 1 回抽選する
func (g *Snap) ReactionMsForTest() int { return g.drawReactionMs() }

// GiveStockForTest は指定席のストックを cards ちょうどに置き換える
func (g *Snap) GiveStockForTest(playerIdx int, cards ...*Card) {
	g.players[playerIdx].ResetStock()
	g.players[playerIdx].AddToStockBottom(cards...)
}
