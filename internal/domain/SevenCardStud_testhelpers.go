//go:build test

package domain

// This file contains test helper methods for SevenCardStud.
// They exist solely for cross-package test setup and are not part of the production game logic.

// SetPhase フェーズ設定（テスト用）
func (s *SevenCardStud) SetPhase(phase int) { s.phase = phase }

// SetCurrentTurn 現在のターン設定（テスト用）
func (s *SevenCardStud) SetCurrentTurn(turn int) { s.currentTurn = turn }

// SetPot ポット設定（テスト用）
func (s *SevenCardStud) SetPot(pot int) { s.pot = pot }

// SetDealerIdx ディーラーインデックス設定（テスト用）
func (s *SevenCardStud) SetDealerIdx(idx int) { s.dealerIdx = idx }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (s *SevenCardStud) SetGameEndFlag(flag bool) { s.gameEndFlag = flag }

// SetLastBet 最後のベット設定（テスト用）
func (s *SevenCardStud) SetLastBet(bet int) { s.lastBet = bet }

// SetMinRaise 最小レイズ額設定（テスト用）
func (s *SevenCardStud) SetMinRaise(raise int) { s.minRaise = raise }

// SetRoundResults ラウンド結果設定（テスト用）
func (s *SevenCardStud) SetRoundResults(results []SevenCardStudResult) { s.roundResults = results }

// SetCpuActions CPU行動記録設定（テスト用）
func (s *SevenCardStud) SetCpuActions(actions []SevenCardStudCpuAction) { s.cpuActions = actions }

// SetSidePots サイドポット設定（テスト用）
func (s *SevenCardStud) SetSidePots(pots []SidePot) { s.sidePots = pots }

// SetHandCount ハンド数設定（テスト用）
func (s *SevenCardStud) SetHandCount(count int) { s.handCount = count }

// SetRebuyCounts リバイ回数設定（テスト用）
func (s *SevenCardStud) SetRebuyCounts(counts []int) { s.rebuyCounts = counts }

// SetAddonUsed アドオン使用フラグ設定（テスト用）
func (s *SevenCardStud) SetAddonUsed(used []bool) { s.addonUsed = used }

// SetRebuyPhaseType リバイフェーズ種別設定（テスト用）
func (s *SevenCardStud) SetRebuyPhaseType(t int) { s.rebuyPhaseType = t }

// SetHumanProfile メタAIプロファイル設定（テスト用）
func (s *SevenCardStud) SetHumanProfile(p *BettingHumanProfile) { s.humanProfile = p }

// SetCommunityCard 共有カード設定（テスト用）
func (s *SevenCardStud) SetCommunityCard(c *Card) { s.communityCard = c }

// SetBringInPlayerIdx ブリングインプレイヤーインデックス設定（テスト用）
func (s *SevenCardStud) SetBringInPlayerIdx(idx int) { s.bringInPlayerIdx = idx }

// SetActedFlags actedフラグ設定（テスト用）
func (s *SevenCardStud) SetActedFlags(flags []bool) { s.actedFlags = flags }

// SetStartingChips 開始チップ設定（テスト用）
func (s *SevenCardStud) SetStartingChips(chips []int) { s.startingChips = chips }

// GetLastHumanPlayMs 最後の人間迷い時間取得（テスト用）
func (s *SevenCardStud) GetLastHumanPlayMs() int { return s.lastHumanPlayMs }

// SetLastHumanPlayMs 最後の人間迷い時間設定（テスト用）
func (s *SevenCardStud) SetLastHumanPlayMs(ms int) { s.lastHumanPlayMs = ms }

// SetRaiseCount レイズ回数設定（テスト用）
func (s *SevenCardStud) SetRaiseCount(count int) { s.raiseCount = count }
