//go:build test

package domain

// This file contains test helper methods for Poker.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetPhase フェーズ設定（テスト用）
func (p *Poker) SetPhase(phase int) { p.phase = phase }

// SetCurrentTurn 現在のターン設定（テスト用）
func (p *Poker) SetCurrentTurn(turn int) { p.currentTurn = turn }

// SetPot ポット設定（テスト用）
func (p *Poker) SetPot(pot int) { p.pot = pot }

// SetDealerIdx ディーラーインデックス設定（テスト用）
func (p *Poker) SetDealerIdx(idx int) { p.dealerIdx = idx }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (p *Poker) SetGameEndFlag(flag bool) { p.gameEndFlag = flag }

// SetLastBet 最後のベット設定（テスト用）
func (p *Poker) SetLastBet(bet int) { p.lastBet = bet }

// SetMinRaise 最小レイズ額設定（テスト用）
func (p *Poker) SetMinRaise(raise int) { p.minRaise = raise }

// SetRoundResults ラウンド結果設定（テスト用）
func (p *Poker) SetRoundResults(results []PokerResult) { p.roundResults = results }

// SetCpuActions CPU行動記録設定（テスト用）
func (p *Poker) SetCpuActions(actions []PokerCpuAction) { p.cpuActions = actions }

// SetCpuExchanges CPU交換記録設定（テスト用）
func (p *Poker) SetCpuExchanges(exchanges []PokerCpuExchange) { p.cpuExchanges = exchanges }

// SetSidePots サイドポット設定（テスト用）
func (p *Poker) SetSidePots(pots []PokerSidePot) { p.sidePots = pots }

// SetHumanProfile メタAIプロファイル設定（テスト用）
func (p *Poker) SetHumanProfile(profile *BettingHumanProfile) { p.humanProfile = profile }

// GetLastHumanPlayMs 最後の人間プレイ時間取得（テスト用）
func (p *Poker) GetLastHumanPlayMs() int { return p.lastHumanPlayMs }
