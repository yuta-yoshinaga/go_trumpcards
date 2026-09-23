//go:build test

package domain

// SetPhaseForTest はフェーズを差し替える (テスト専用)。
func (g *Guandan) SetPhaseForTest(p GuandanPhase) { g.phase = p }

// SetHandForTest は手札を差し替える (テスト専用)。
func (g *Guandan) SetHandForTest(idx int, cards []*Card) {
	setHandForTest(g.GetPlayer(idx), cards)
}

// SetLevelForTest は基準レベルを差し替える (テスト専用)。
func (g *Guandan) SetLevelForTest(level int) { g.level = level }

// SetTeamLevelForTest はチームのレベルを差し替える (テスト専用)。
func (g *Guandan) SetTeamLevelForTest(team, level int) {
	if team >= 0 && team < GuandanTeamCnt {
		g.levels[team] = level
	}
}

// SetCurrentPlayerForTest は手番を差し替える (テスト専用)。
func (g *Guandan) SetCurrentPlayerForTest(idx int) { g.currentIdx = idx }

// SetFinishedForTest は上がり順を差し替える (テスト専用)。
func (g *Guandan) SetFinishedForTest(order []int) { g.finished = order }

// FinishHandForTest は精算を走らせる (テスト専用)。
func (g *Guandan) FinishHandForTest() { g.finishHand() }

// PrepareTributeForTest は進貢を走らせる (テスト専用)。
func (g *Guandan) PrepareTributeForTest(prev *GuandanHandResult) { g.prepareTribute(prev) }
