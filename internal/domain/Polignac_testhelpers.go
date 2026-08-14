//go:build test

package domain

// This file contains test helper methods for Polignac.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetPhaseForTest フェーズを設定する（テスト用）
func (p *Polignac) SetPhaseForTest(phase PolignacPhase) { p.phase = phase }

// SetTrickNumberForTest 現在のトリック番号を設定する（テスト用）
func (p *Polignac) SetTrickNumberForTest(n int) { p.trickNumber = n }

// SetCurrentPlayerIdxForTest 手番を設定する（テスト用）
func (p *Polignac) SetCurrentPlayerIdxForTest(i int) { p.currentPlayerIdx = i }

// SetCurrentTrickForTest 場に出ている札を設定する（テスト用）
func (p *Polignac) SetCurrentTrickForTest(t []*TrickCard) { p.currentTrick = t }

// SetCapotIdxForTest capot 宣言者を設定する（テスト用）
func (p *Polignac) SetCapotIdxForTest(i int) { p.capotIdx = i }

// SetCapotTricksForTest capot 宣言者の獲得トリック数を設定する（テスト用）
func (p *Polignac) SetCapotTricksForTest(n int) { p.capotTricks = n }

// FinishRoundForTest ラウンドの失点を確定させる（テスト用）
func (p *Polignac) FinishRoundForTest() { p.finishRound() }

// FinishGameForTest 現在の累計失点で勝敗を確定させる（テスト用）
func (p *Polignac) FinishGameForTest() { p.finishGame() }
