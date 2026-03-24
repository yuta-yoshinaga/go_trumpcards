//go:build test

package domain

// This file contains test helper methods for ShortDeck.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetPhase フェーズ設定（テスト用）
func (sd *ShortDeck) SetPhase(phase int) { sd.phase = phase }

// SetCurrentTurn 現在のターン設定（テスト用）
func (sd *ShortDeck) SetCurrentTurn(turn int) { sd.currentTurn = turn }

// SetCommunityCards コミュニティカード設定（テスト用）
func (sd *ShortDeck) SetCommunityCards(cards []*Card) { sd.communityCards = cards }

// SetPot ポット設定（テスト用）
func (sd *ShortDeck) SetPot(pot int) { sd.pot = pot }

// SetDealerIdx ディーラーインデックス設定（テスト用）
func (sd *ShortDeck) SetDealerIdx(idx int) { sd.dealerIdx = idx }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (sd *ShortDeck) SetGameEndFlag(flag bool) { sd.gameEndFlag = flag }

// SetLastBet 最後のベット設定（テスト用）
func (sd *ShortDeck) SetLastBet(bet int) { sd.lastBet = bet }

// SetMinRaise 最小レイズ額設定（テスト用）
func (sd *ShortDeck) SetMinRaise(raise int) { sd.minRaise = raise }

// SetRoundResults ラウンド結果設定（テスト用）
func (sd *ShortDeck) SetRoundResults(results []ShortDeckResult) { sd.roundResults = results }

// SetCpuActions CPU行動記録設定（テスト用）
func (sd *ShortDeck) SetCpuActions(actions []ShortDeckCpuAction) { sd.cpuActions = actions }

// SetSidePots サイドポット設定（テスト用）
func (sd *ShortDeck) SetSidePots(pots []ShortDeckSidePot) { sd.sidePots = pots }

// SetHandCount ハンド数設定（テスト用）
func (sd *ShortDeck) SetHandCount(count int) { sd.handCount = count }

// SetRebuyCounts リバイ回数設定（テスト用）
func (sd *ShortDeck) SetRebuyCounts(counts []int) { sd.rebuyCounts = counts }

// SetAddonUsed アドオン使用フラグ設定（テスト用）
func (sd *ShortDeck) SetAddonUsed(used []bool) { sd.addonUsed = used }

// SetRebuyPhaseType リバイフェーズ種別設定（テスト用）
func (sd *ShortDeck) SetRebuyPhaseType(t int) { sd.rebuyPhaseType = t }

// SetHumanProfile メタAIプロファイル設定（テスト用）
func (sd *ShortDeck) SetHumanProfile(profile *BettingHumanProfile) { sd.humanProfile = profile }

// GetLastHumanPlayMs 最後の人間プレイ時間取得（テスト用）
func (sd *ShortDeck) GetLastHumanPlayMs() int { return sd.lastHumanPlayMs }

// SetActedFlags actedフラグ設定（テスト用）
func (sd *ShortDeck) SetActedFlags(flags []bool) { sd.actedFlags = flags }

// SetRaiseCount レイズ回数設定（テスト用）
func (sd *ShortDeck) SetRaiseCount(count int) { sd.raiseCount = count }

// SetStartingChips 開始時チップ設定（テスト用）
func (sd *ShortDeck) SetStartingChips(chips []int) { sd.startingChips = chips }
