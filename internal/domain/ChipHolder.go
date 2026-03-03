package domain

// ChipHolder チップ管理の共通構造体
type ChipHolder struct {
	chips int
}

// GetChips チップ取得
func (ch *ChipHolder) GetChips() int {
	return ch.chips
}

// SetChips チップ設定
func (ch *ChipHolder) SetChips(chips int) {
	ch.chips = chips
}

// AddChips チップ追加
func (ch *ChipHolder) AddChips(amount int) {
	ch.chips += amount
}

// SubtractChips チップ減算 (不足時はfalseを返す)
func (ch *ChipHolder) SubtractChips(amount int) bool {
	if ch.chips < amount {
		return false
	}
	ch.chips -= amount
	return true
}
