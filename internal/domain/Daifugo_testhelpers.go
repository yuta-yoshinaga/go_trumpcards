//go:build test

package domain

// This file contains test helper methods for Daifugo.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetElevenBackActive 11バック設定（テスト用）
func (d *Daifugo) SetElevenBackActive(active bool) { d.round.elevenBackActive = active }

// SetSuitLocked スート縛り設定（テスト用）
func (d *Daifugo) SetSuitLocked(locked bool, suit int) {
	d.round.suitLocked = locked
	d.round.lockedSuit = suit
}

// SetTableIsSequence 階段フラグ設定（テスト用）
func (d *Daifugo) SetTableIsSequence(seq bool) { d.round.tableIsSequence = seq }

// SetExchangeActions カード交換記録設定（テスト用）
func (d *Daifugo) SetExchangeActions(actions []*DaifugoExchangeAction) {
	d.round.exchangeActions = actions
}

// SetHumanAction 人間の行動設定（テスト用）
func (d *Daifugo) SetHumanAction(action *DaifugoCpuAction) { d.round.humanAction = action }

// SetReverseDirection 9リバースの方向設定（テスト用）
func (d *Daifugo) SetReverseDirection(v bool) { d.round.reverseDirection = v }

// SetNumberLocked 連番縛り設定（テスト用）
func (d *Daifugo) SetNumberLocked(v bool) { d.round.numberLocked = v }

// SetSortMode 手札ソートモード設定（テスト用）
func (d *Daifugo) SetSortMode(mode DaifugoSortMode) { d.sortMode = mode }

// SetTableCards 場札設定（テスト用）
func (d *Daifugo) SetTableCards(cards []*Card) { d.round.tableCards = cards }

// SetCurrentTurn ターン設定（テスト用）
func (d *Daifugo) SetCurrentTurn(turn int) { d.round.currentTurn = turn }

// SetLastPlayPlayerIdx 最後にカードを出したプレイヤー設定（テスト用）
func (d *Daifugo) SetLastPlayPlayerIdx(idx int) { d.round.lastPlayPlayerIdx = idx }

// SetRevolutionActive 革命フラグ設定（テスト用）
func (d *Daifugo) SetRevolutionActive(v bool) { d.round.revolutionActive = v }

// SetSequenceLocked 階段縛り設定（テスト用）
func (d *Daifugo) SetSequenceLocked(v bool) { d.round.sequenceLocked = v }
