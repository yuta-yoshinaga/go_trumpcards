package domain

// This file contains test helper methods for Sevens.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetHumanAction 人間の行動設定（テスト用）
func (s *Sevens) SetHumanAction(action *SevensCpuAction) { s.humanAction = action }

// SetCpuActions CPU行動設定（テスト用）
func (s *Sevens) SetCpuActions(actions []*SevensCpuAction) { s.cpuActions = actions }
