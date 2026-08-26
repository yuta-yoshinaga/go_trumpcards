//go:build test

package domain

// This file contains test helper methods for Julepe.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetPhaseForTest フェーズを設定する（テスト用）
func (r *Julepe) SetPhaseForTest(phase JulepePhase) { r.phase = phase }

// SetTrickNumberForTest 現在のトリック番号を設定する（テスト用）
func (r *Julepe) SetTrickNumberForTest(n int) { r.trickNumber = n }

// SetCurrentPlayerIdxForTest 手番を設定する（テスト用）
func (r *Julepe) SetCurrentPlayerIdxForTest(i int) { r.currentPlayerIdx = i }

// SetCurrentTrickForTest 場に出ている札を設定する（テスト用）
func (r *Julepe) SetCurrentTrickForTest(t []*TrickCard) { r.currentTrick = t }

// SetPotForTest ポットのチップを設定する（テスト用）
func (r *Julepe) SetPotForTest(n int) { r.pot = n }

// SetTrumpSuitForTest 切り札を設定する（テスト用）
func (r *Julepe) SetTrumpSuitForTest(suit int) { r.trumpSuit = suit }

// StartPlayIfReadyForTest 全員の選択が済んでいればプレイに入る（テスト用）
func (r *Julepe) StartPlayIfReadyForTest() { r.startPlayIfReady() }

// FinishRoundForTest ラウンドの配当を確定させる（テスト用）
func (r *Julepe) FinishRoundForTest() { r.finishRound() }

// FinishGameForTest 現在のチップで勝敗を確定させる（テスト用）
func (r *Julepe) FinishGameForTest() { r.finishGame() }

// SetDealerIdxForTest は親の席を差し替える（テスト用）。
func (r *Julepe) SetDealerIdxForTest(i int) { r.dealerIdx = i }
