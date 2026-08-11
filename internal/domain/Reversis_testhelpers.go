//go:build test

package domain

// This file contains test helper methods for Reversis.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetPhaseForTest フェーズを設定する（テスト用）
func (r *Reversis) SetPhaseForTest(phase ReversisPhase) { r.phase = phase }

// SetTrickNumberForTest 現在のトリック番号を設定する（テスト用）
func (r *Reversis) SetTrickNumberForTest(n int) { r.trickNumber = n }

// SetCurrentPlayerIdxForTest 手番を設定する（テスト用）
func (r *Reversis) SetCurrentPlayerIdxForTest(i int) { r.currentPlayerIdx = i }

// SetCurrentTrickForTest 場に出ている札を設定する（テスト用）
func (r *Reversis) SetCurrentTrickForTest(t []*TrickCard) { r.currentTrick = t }

// SetPoolForTest プールのチップを設定する（テスト用）
func (r *Reversis) SetPoolForTest(n int) { r.pool = n }

// FinishRoundForTest ラウンドの配当を確定させる（テスト用）
func (r *Reversis) FinishRoundForTest() { r.finishRound() }

// FinishGameForTest 現在のチップで勝敗を確定させる（テスト用）
func (r *Reversis) FinishGameForTest() { r.finishGame() }
