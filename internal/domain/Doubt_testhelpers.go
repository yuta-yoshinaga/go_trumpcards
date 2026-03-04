package domain

// This file contains test helper methods for Doubt.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetPhase フェーズ設定（テスト用）
func (d *Doubt) SetPhase(phase DoubtPhase) { d.phase = phase }

// SetLastAction 最後のプレイアクション設定（テスト用）
func (d *Doubt) SetLastAction(action *DoubtAction) { d.lastAction = action }

// SetCpuDoubters CPUダウターリスト設定（テスト用）
func (d *Doubt) SetCpuDoubters(doubters []int) { d.cpuDoubters = doubters }

// SetCpuActions CPU行動設定（テスト用）
func (d *Doubt) SetCpuActions(actions []*DoubtCpuAction) { d.cpuActions = actions }

// SetHumanAction 人間の行動設定（テスト用）
func (d *Doubt) SetHumanAction(action *DoubtCpuAction) { d.humanAction = action }

// SetLastDoubtResult ダウト結果設定（テスト用）
func (d *Doubt) SetLastDoubtResult(result *DoubtDoubtResult) { d.lastDoubtResult = result }

// SetWinnerIdx 勝者インデックス設定（テスト用）
func (d *Doubt) SetWinnerIdx(idx int) { d.winnerIdx = idx }
