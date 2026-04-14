//go:build test

package domain

// This file contains test helper methods for Omaha.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetPhase フェーズ設定（テスト用）
func (o *Omaha) SetPhase(phase int) { o.phase = phase }

// SetCurrentTurn 現在のターン設定（テスト用）
func (o *Omaha) SetCurrentTurn(turn int) { o.currentTurn = turn }

// SetCommunityCards コミュニティカード設定（テスト用）
func (o *Omaha) SetCommunityCards(cards []*Card) { o.communityCards = cards }

// SetPot ポット設定（テスト用）
func (o *Omaha) SetPot(pot int) { o.pot = pot }

// SetDealerIdx ディーラーインデックス設定（テスト用）
func (o *Omaha) SetDealerIdx(idx int) { o.dealerIdx = idx }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (o *Omaha) SetGameEndFlag(flag bool) { o.gameEndFlag = flag }

// SetLastBet 最後のベット設定（テスト用）
func (o *Omaha) SetLastBet(bet int) { o.lastBet = bet }

// SetMinRaise 最小レイズ額設定（テスト用）
func (o *Omaha) SetMinRaise(raise int) { o.minRaise = raise }

// SetRoundResults ラウンド結果設定（テスト用）
func (o *Omaha) SetRoundResults(results []HoldemResult) { o.roundResults = results }

// SetCpuActions CPU行動記録設定（テスト用）
func (o *Omaha) SetCpuActions(actions []HoldemCpuAction) { o.cpuActions = actions }

// SetSidePots サイドポット設定（テスト用）
func (o *Omaha) SetSidePots(pots []SidePot) { o.sidePots = pots }

// SetHandCount ハンド数設定（テスト用）
func (o *Omaha) SetHandCount(count int) { o.handCount = count }

// SetRebuyCounts リバイ回数設定（テスト用）
func (o *Omaha) SetRebuyCounts(counts []int) { o.rebuyCounts = counts }

// SetAddonUsed アドオン使用フラグ設定（テスト用）
func (o *Omaha) SetAddonUsed(used []bool) { o.addonUsed = used }

// SetRebuyPhaseType リバイフェーズ種別設定（テスト用）
func (o *Omaha) SetRebuyPhaseType(t int) { o.rebuyPhaseType = t }

// SetHumanProfile メタAIプロファイル設定（テスト用）
func (o *Omaha) SetHumanProfile(profile *BettingHumanProfile) { o.humanProfile = profile }

// GetLastHumanPlayMs 最後の人間プレイ時間取得（テスト用）
func (o *Omaha) GetLastHumanPlayMs() int { return o.lastHumanPlayMs }
