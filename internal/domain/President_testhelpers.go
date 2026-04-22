//go:build test

package domain

// Test helper methods for President. They exist solely for cross-package test setup
// (e.g. presenter tests) and are not part of the production game logic.

// SetTableCards 場札設定 (テスト用)
func (p *President) SetTableCards(cards []*Card) { p.round.tableCards = cards }

// SetCurrentTurn ターン設定 (テスト用)
func (p *President) SetCurrentTurn(turn int) { p.round.currentTurn = turn }

// SetLastPlayPlayerIdx 最後に出したプレイヤー設定 (テスト用)
func (p *President) SetLastPlayPlayerIdx(idx int) { p.round.lastPlayPlayerIdx = idx }

// SetRevolutionActive 革命フラグ設定 (テスト用)
func (p *President) SetRevolutionActive(v bool) { p.round.revolutionActive = v }

// SetExchangeActions カード交換記録設定 (テスト用)
func (p *President) SetExchangeActions(actions []*PresidentExchangeAction) {
	p.round.exchangeActions = actions
}

// SetHumanAction 人間の行動設定 (テスト用)
func (p *President) SetHumanAction(action *PresidentCpuAction) { p.round.humanAction = action }
