//go:build test

package domain

// This file contains test helper methods for OldMaid.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetLastDrawPlayerIdx 最後に引いたプレイヤーインデックス設定（テスト用）
func (o *OldMaid) SetLastDrawPlayerIdx(idx int) { o.lastDrawPlayerIdx = idx }

// SetHasDrawn 引いたフラグ設定（テスト用）
func (o *OldMaid) SetHasDrawn(v bool) { o.hasDrawn = v }

// SetHumanAction 人間プレイヤーの行動設定（テスト用）
func (o *OldMaid) SetHumanAction(action *OldMaidCpuAction) { o.humanAction = action }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (o *OldMaid) SetGameEndFlag(v bool) { o.gameEndFlag = v }

// SetCpuHighlightedCardIdx CPU心理戦の強調インデックス設定（テスト用）
func (o *OldMaid) SetCpuHighlightedCardIdx(v int) { o.cpuHighlightedCardIdx = v }

// SetRemovedCard 除外カード設定（テスト用）
func (o *OldMaid) SetRemovedCard(card *Card) { o.removedCard = card }

// SetDrawHistory 引き履歴設定（テスト用）
func (o *OldMaid) SetDrawHistory(h []*OldMaidDrawHistoryEntry) { o.drawHistory = h }
