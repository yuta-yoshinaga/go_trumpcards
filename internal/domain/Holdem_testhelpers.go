//go:build test

package domain

// This file contains test helper methods for Holdem.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetPhase フェーズ設定（テスト用）
func (h *Holdem) SetPhase(phase int) { h.phase = phase }

// SetCurrentTurn 現在のターン設定（テスト用）
func (h *Holdem) SetCurrentTurn(turn int) { h.currentTurn = turn }

// SetCommunityCards コミュニティカード設定（テスト用）
func (h *Holdem) SetCommunityCards(cards []*Card) { h.communityCards = cards }

// SetPot ポット設定（テスト用）
func (h *Holdem) SetPot(pot int) { h.pot = pot }

// SetDealerIdx ディーラーインデックス設定（テスト用）
func (h *Holdem) SetDealerIdx(idx int) { h.dealerIdx = idx }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (h *Holdem) SetGameEndFlag(flag bool) { h.gameEndFlag = flag }

// SetLastBet 最後のベット設定（テスト用）
func (h *Holdem) SetLastBet(bet int) { h.lastBet = bet }

// SetMinRaise 最小レイズ額設定（テスト用）
func (h *Holdem) SetMinRaise(raise int) { h.minRaise = raise }

// SetRoundResults ラウンド結果設定（テスト用）
func (h *Holdem) SetRoundResults(results []HoldemResult) { h.roundResults = results }

// SetCpuActions CPU行動記録設定（テスト用）
func (h *Holdem) SetCpuActions(actions []HoldemCpuAction) { h.cpuActions = actions }

// SetSidePots サイドポット設定（テスト用）
func (h *Holdem) SetSidePots(pots []HoldemSidePot) { h.sidePots = pots }

// SetHandCount ハンド数設定（テスト用）
func (h *Holdem) SetHandCount(count int) { h.handCount = count }

// SetRebuyCounts リバイ回数設定（テスト用）
func (h *Holdem) SetRebuyCounts(counts []int) { h.rebuyCounts = counts }

// SetAddonUsed アドオン使用フラグ設定（テスト用）
func (h *Holdem) SetAddonUsed(used []bool) { h.addonUsed = used }

// SetRebuyPhaseType リバイフェーズ種別設定（テスト用）
func (h *Holdem) SetRebuyPhaseType(t int) { h.rebuyPhaseType = t }

// SetHumanProfile メタAIプロファイル設定（テスト用）
func (h *Holdem) SetHumanProfile(profile *BettingHumanProfile) { h.humanProfile = profile }

// GetLastHumanPlayMs 最後の人間プレイ時間取得（テスト用）
func (h *Holdem) GetLastHumanPlayMs() int { return h.lastHumanPlayMs }
