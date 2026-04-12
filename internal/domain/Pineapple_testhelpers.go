//go:build test

package domain

// This file contains test helper methods for Pineapple.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetPhase フェーズ設定（テスト用）
func (p *Pineapple) SetPhase(phase int) { p.phase = phase }

// SetCurrentTurn 現在のターン設定（テスト用）
func (p *Pineapple) SetCurrentTurn(turn int) { p.currentTurn = turn }

// SetCommunityCards コミュニティカード設定（テスト用）
func (p *Pineapple) SetCommunityCards(cards []*Card) { p.communityCards = cards }

// SetPot ポット設定（テスト用）
func (p *Pineapple) SetPot(pot int) { p.pot = pot }

// SetDealerIdx ディーラーインデックス設定（テスト用）
func (p *Pineapple) SetDealerIdx(idx int) { p.dealerIdx = idx }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (p *Pineapple) SetGameEndFlag(flag bool) { p.gameEndFlag = flag }

// SetLastBet 最後のベット設定（テスト用）
func (p *Pineapple) SetLastBet(bet int) { p.lastBet = bet }

// SetMinRaise 最小レイズ額設定（テスト用）
func (p *Pineapple) SetMinRaise(raise int) { p.minRaise = raise }

// SetRoundResults ラウンド結果設定（テスト用）
func (p *Pineapple) SetRoundResults(results []PineappleResult) { p.roundResults = results }

// SetCpuActions CPU行動記録設定（テスト用）
func (p *Pineapple) SetCpuActions(actions []PineappleCpuAction) { p.cpuActions = actions }

// SetSidePots サイドポット設定（テスト用）
func (p *Pineapple) SetSidePots(pots []PineappleSidePot) { p.sidePots = pots }

// SetHandCount ハンド数設定（テスト用）
func (p *Pineapple) SetHandCount(count int) { p.handCount = count }

// SetRebuyCounts リバイ回数設定（テスト用）
func (p *Pineapple) SetRebuyCounts(counts []int) { p.rebuyCounts = counts }

// SetAddonUsed アドオン使用フラグ設定（テスト用）
func (p *Pineapple) SetAddonUsed(used []bool) { p.addonUsed = used }

// SetRebuyPhaseType リバイフェーズ種別設定（テスト用）
func (p *Pineapple) SetRebuyPhaseType(t int) { p.rebuyPhaseType = t }

// SetHumanProfile メタAIプロファイル設定（テスト用）
func (p *Pineapple) SetHumanProfile(profile *BettingHumanProfile) { p.humanProfile = profile }

// GetLastHumanPlayMs 最後の人間プレイ時間取得（テスト用）
func (p *Pineapple) GetLastHumanPlayMs() int { return p.lastHumanPlayMs }
