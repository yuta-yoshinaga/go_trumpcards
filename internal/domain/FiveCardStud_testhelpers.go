//go:build test

package domain

// This file contains test helper methods for FiveCardStud.
// They exist solely for cross-package test setup and are not part of the production game logic.

// SetPhase フェーズ設定（テスト用）
func (s *FiveCardStud) SetPhase(phase int) { s.phase = phase }

// SetCurrentTurn 現在のターン設定（テスト用）
func (s *FiveCardStud) SetCurrentTurn(turn int) { s.currentTurn = turn }

// SetPot ポット設定（テスト用）
func (s *FiveCardStud) SetPot(pot int) { s.pot = pot }

// SetDealerIdx ディーラーインデックス設定（テスト用）
func (s *FiveCardStud) SetDealerIdx(idx int) { s.dealerIdx = idx }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (s *FiveCardStud) SetGameEndFlag(flag bool) { s.gameEndFlag = flag }

// SetLastBet 最後のベット設定（テスト用）
func (s *FiveCardStud) SetLastBet(bet int) { s.lastBet = bet }

// SetMinRaise 最小レイズ額設定（テスト用）
func (s *FiveCardStud) SetMinRaise(raise int) { s.minRaise = raise }

// SetRoundResults ラウンド結果設定（テスト用）
func (s *FiveCardStud) SetRoundResults(results []FiveCardStudResult) { s.roundResults = results }

// SetCpuActions CPU行動記録設定（テスト用）
func (s *FiveCardStud) SetCpuActions(actions []FiveCardStudCpuAction) { s.cpuActions = actions }

// SetSidePots サイドポット設定（テスト用）
func (s *FiveCardStud) SetSidePots(pots []SidePot) { s.sidePots = pots }

// SetHandCount ハンド数設定（テスト用）
func (s *FiveCardStud) SetHandCount(count int) { s.handCount = count }

// SetRebuyCounts リバイ回数設定（テスト用）
func (s *FiveCardStud) SetRebuyCounts(counts []int) { s.rebuyCounts = counts }

// SetAddonUsed アドオン使用フラグ設定（テスト用）
func (s *FiveCardStud) SetAddonUsed(used []bool) { s.addonUsed = used }

// SetRebuyPhaseType リバイフェーズ種別設定（テスト用）
func (s *FiveCardStud) SetRebuyPhaseType(t int) { s.rebuyPhaseType = t }

// SetHumanProfile メタAIプロファイル設定（テスト用）
func (s *FiveCardStud) SetHumanProfile(p *BettingHumanProfile) { s.humanProfile = p }

// SetCommunityCard 共有カード設定（テスト用）
func (s *FiveCardStud) SetCommunityCard(c *Card) { s.communityCard = c }

// SetBringInPlayerIdx ブリングインプレイヤーインデックス設定（テスト用）
func (s *FiveCardStud) SetBringInPlayerIdx(idx int) { s.bringInPlayerIdx = idx }

// SetActedFlags actedフラグ設定（テスト用）
func (s *FiveCardStud) SetActedFlags(flags []bool) { s.actedFlags = flags }

// SetStartingChips 開始チップ設定（テスト用）
func (s *FiveCardStud) SetStartingChips(chips []int) { s.startingChips = chips }

// GetLastHumanPlayMs 最後の人間迷い時間取得（テスト用）
func (s *FiveCardStud) GetLastHumanPlayMs() int { return s.lastHumanPlayMs }

// SetLastHumanPlayMs 最後の人間迷い時間設定（テスト用）
func (s *FiveCardStud) SetLastHumanPlayMs(ms int) { s.lastHumanPlayMs = ms }

// SetRaiseCount レイズ回数設定（テスト用）
func (s *FiveCardStud) SetRaiseCount(count int) { s.raiseCount = count }
