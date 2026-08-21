//go:build test

package domain

// This file contains test helper methods for FollowTheQueen.
// They exist solely for cross-package test setup and are not part of the production game logic.

// SetPhase フェーズ設定（テスト用）
func (s *FollowTheQueen) SetPhase(phase int) { s.phase = phase }

// SetCurrentTurn 現在のターン設定（テスト用）
func (s *FollowTheQueen) SetCurrentTurn(turn int) { s.currentTurn = turn }

// SetPot ポット設定（テスト用）
func (s *FollowTheQueen) SetPot(pot int) { s.pot = pot }

// SetDealerIdx ディーラーインデックス設定（テスト用）
func (s *FollowTheQueen) SetDealerIdx(idx int) { s.dealerIdx = idx }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (s *FollowTheQueen) SetGameEndFlag(flag bool) { s.gameEndFlag = flag }

// SetLastBet 最後のベット設定（テスト用）
func (s *FollowTheQueen) SetLastBet(bet int) { s.lastBet = bet }

// SetMinRaise 最小レイズ額設定（テスト用）
func (s *FollowTheQueen) SetMinRaise(raise int) { s.minRaise = raise }

// SetRoundResults ラウンド結果設定（テスト用）
func (s *FollowTheQueen) SetRoundResults(results []FollowTheQueenResult) { s.roundResults = results }

// SetCpuActions CPU行動記録設定（テスト用）
func (s *FollowTheQueen) SetCpuActions(actions []FollowTheQueenCpuAction) { s.cpuActions = actions }

// SetSidePots サイドポット設定（テスト用）
func (s *FollowTheQueen) SetSidePots(pots []SidePot) { s.sidePots = pots }

// SetHandCount ハンド数設定（テスト用）
func (s *FollowTheQueen) SetHandCount(count int) { s.handCount = count }

// SetRebuyCounts リバイ回数設定（テスト用）
func (s *FollowTheQueen) SetRebuyCounts(counts []int) { s.rebuyCounts = counts }

// SetAddonUsed アドオン使用フラグ設定（テスト用）
func (s *FollowTheQueen) SetAddonUsed(used []bool) { s.addonUsed = used }

// SetRebuyPhaseType リバイフェーズ種別設定（テスト用）
func (s *FollowTheQueen) SetRebuyPhaseType(t int) { s.rebuyPhaseType = t }

// SetHumanProfile メタAIプロファイル設定（テスト用）
func (s *FollowTheQueen) SetHumanProfile(p *BettingHumanProfile) { s.humanProfile = p }

// SetCommunityCard 共有カード設定（テスト用）
func (s *FollowTheQueen) SetCommunityCard(c *Card) { s.communityCard = c }

// SetBringInPlayerIdx ブリングインプレイヤーインデックス設定（テスト用）
func (s *FollowTheQueen) SetBringInPlayerIdx(idx int) { s.bringInPlayerIdx = idx }

// SetActedFlags actedフラグ設定（テスト用）
func (s *FollowTheQueen) SetActedFlags(flags []bool) { s.actedFlags = flags }

// SetStartingChips 開始チップ設定（テスト用）
func (s *FollowTheQueen) SetStartingChips(chips []int) { s.startingChips = chips }

// GetLastHumanPlayMs 最後の人間迷い時間取得（テスト用）
func (s *FollowTheQueen) GetLastHumanPlayMs() int { return s.lastHumanPlayMs }

// SetLastHumanPlayMs 最後の人間迷い時間設定（テスト用）
func (s *FollowTheQueen) SetLastHumanPlayMs(ms int) { s.lastHumanPlayMs = ms }

// SetRaiseCount レイズ回数設定（テスト用）
func (s *FollowTheQueen) SetRaiseCount(count int) { s.raiseCount = count }

// SetWildRankForTest はワイルドランクを直接設定する（テスト用）。
//
// 本番では noteUpCard がディールの過程で設定するので、外から狙った値にする手段が
// 無い。配りに依存したテストを書かないための helper。
func (s *FollowTheQueen) SetWildRankForTest(rank int) {
	s.wildRank = rank
	s.publishWildRank()
}
