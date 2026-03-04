package domain

// setActedFlags actedフラグ設定（テスト用）
func (p *Poker) setActedFlags(flags []bool) { p.actedFlags = flags }

// setRaiseCount レイズ回数設定（テスト用）
func (p *Poker) setRaiseCount(count int) { p.raiseCount = count }

// setStartingChips ハンド開始時チップ設定（テスト用）
func (p *Poker) setStartingChips(chips []int) { p.startingChips = chips }

// getStartingChips ハンド開始時チップ取得（テスト用）
func (p *Poker) getStartingChips() []int { return p.startingChips }

// getActedFlags actedフラグ取得（テスト用）
func (p *Poker) getActedFlags() []bool {
	result := make([]bool, len(p.actedFlags))
	copy(result, p.actedFlags)
	return result
}
