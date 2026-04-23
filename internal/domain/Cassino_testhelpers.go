//go:build test

package domain

// Test helper methods for Cassino. They exist solely for cross-package test setup
// (e.g. presenter tests) and are not part of the production game logic.

// SetTableCards 場札設定 (テスト用)
func (c *Cassino) SetTableCards(cards []*Card) { c.round.tableCards = cards }

// SetBuilds ビルド設定 (テスト用)
func (c *Cassino) SetBuilds(builds []*CassinoBuild) { c.round.builds = builds }

// SetCurrentTurn ターン設定 (テスト用)
func (c *Cassino) SetCurrentTurn(turn int) { c.round.currentTurn = turn }

// SetLastCaptureIdx 最後の捕獲者 (テスト用)
func (c *Cassino) SetLastCaptureIdx(idx int) { c.round.lastCaptureIdx = idx }

// SetPhase フェーズ設定 (テスト用)
func (c *Cassino) SetPhase(phase string) { c.round.phase = phase }

// SetGameEndFlag ゲーム終了フラグ (テスト用)
func (c *Cassino) SetGameEndFlag(flag bool) { c.round.gameEndFlag = flag }

// SetHumanAction 人間の行動設定 (テスト用)
func (c *Cassino) SetHumanAction(action *CassinoAction) { c.round.humanAction = action }

// SetCpuActions CPU の行動設定 (テスト用)
func (c *Cassino) SetCpuActions(actions []*CassinoAction) { c.round.cpuActions = actions }

// ApplyTakeForTest はテスト専用の applyTake ラッパー (プレイヤー指定可)。
func (c *Cassino) ApplyTakeForTest(playerIdx, handIdx int, tableIdxs, buildIdxs []int) error {
	return c.applyTake(playerIdx, handIdx, tableIdxs, buildIdxs, func(a *CassinoAction) {
		c.round.humanAction = a
	})
}

// ApplyBuildForTest はテスト専用の applyBuild ラッパー (プレイヤー指定可)。
func (c *Cassino) ApplyBuildForTest(playerIdx, handIdx int, tableIdxs []int, declared int) error {
	return c.applyBuild(playerIdx, handIdx, tableIdxs, declared, func(a *CassinoAction) {
		c.round.humanAction = a
	})
}

// ApplyTrailForTest はテスト専用の applyTrail ラッパー (プレイヤー指定可)。
func (c *Cassino) ApplyTrailForTest(playerIdx, handIdx int) error {
	return c.applyTrail(playerIdx, handIdx, func(a *CassinoAction) {
		c.round.humanAction = a
	})
}

// FinishRoundForTest はテスト専用のラウンド終了呼び出し。
func (c *Cassino) FinishRoundForTest() { c.finishRound() }

// ScoreRoundForTest はテスト専用の scoreRound 呼び出し。
func (c *Cassino) ScoreRoundForTest() *CassinoScoreDetail { return c.scoreRound() }

// DealNextPackForTest はテスト専用の dealNextPack。
func (c *Cassino) DealNextPackForTest() { c.dealNextPack() }
