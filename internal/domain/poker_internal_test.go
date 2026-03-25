package domain

// setActedFlags actedフラグ設定（テスト用）
func (p *Poker) setActedFlags(flags []bool) { p.round.actedFlags = flags }

// setRaiseCount レイズ回数設定（テスト用）
func (p *Poker) setRaiseCount(count int) { p.round.raiseCount = count }

// setStartingChips ハンド開始時チップ設定（テスト用）
func (p *Poker) setStartingChips(chips []int) { p.round.startingChips = chips }

// getStartingChips ハンド開始時チップ取得（テスト用）
func (p *Poker) getStartingChips() []int { return p.round.startingChips }

// getActedFlags actedフラグ取得（テスト用）
func (p *Poker) getActedFlags() []bool {
	result := make([]bool, len(p.round.actedFlags))
	copy(result, p.round.actedFlags)
	return result
}
