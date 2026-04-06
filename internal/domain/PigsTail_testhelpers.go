//go:build test

package domain

// This file contains test helper methods for PigsTail.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (pt *PigsTail) SetGameEndFlag(v bool) { pt.gameEndFlag = v }

// SetLoserIdx 負けプレイヤーインデックス設定（テスト用）
func (pt *PigsTail) SetLoserIdx(idx int) { pt.loserIdx = idx }

// SetCurrentTurn 手番設定（テスト用）
func (pt *PigsTail) SetCurrentTurn(turn int) { pt.currentTurn = turn }

// SetLastDrawCard 最後に引いたカード設定（テスト用）
func (pt *PigsTail) SetLastDrawCard(card *Card) { pt.lastDrawCard = card }

// SetLastPenalty 最後のペナルティフラグ設定（テスト用）
func (pt *PigsTail) SetLastPenalty(v bool) { pt.lastPenalty = v }

// SetCenter 場札設定（テスト用）
func (pt *PigsTail) SetCenter(cards []*Card) { pt.center = cards }

// SetCpuActions CPU行動設定（テスト用）
func (pt *PigsTail) SetCpuActions(actions []*PigsTailCpuAction) { pt.cpuActions = actions }

// SetHumanAction 人間の行動設定（テスト用）
func (pt *PigsTail) SetHumanAction(action *PigsTailCpuAction) { pt.humanAction = action }

// SetTrumpCards 山札設定（テスト用）
func (pt *PigsTail) SetTrumpCards(tc *TrumpCards) { pt.trumpCards = tc }
