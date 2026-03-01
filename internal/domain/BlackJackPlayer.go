package domain

// BlackJackPlayer ブラックジャックプレイヤークラス
type BlackJackPlayer struct {
	Player     // 親クラス
	chips  int // チップ
}

// NewBlackJackPlayer コンストラクタ
func NewBlackJackPlayer() *BlackJackPlayer {
	return &BlackJackPlayer{
		Player: Player{
			cards: make([]*Card, 0),
		},
		chips: 0,
	}
}

// AddCard カード追加
func (bp *BlackJackPlayer) AddCard(card *Card) {
	bp.cards = append(bp.cards, card)
}

// GetScore 手札から現在のスコア計算
func (bp *BlackJackPlayer) GetScore() int {
	return CalculateBlackJackScore(bp.cards)
}

// GetChips チップ取得
func (bp *BlackJackPlayer) GetChips() int {
	return bp.chips
}

// SetChips チップ設定
func (bp *BlackJackPlayer) SetChips(chips int) {
	bp.chips = chips
}

// AddChips チップ追加
func (bp *BlackJackPlayer) AddChips(amount int) {
	bp.chips += amount
}

// SubtractChips チップ減算 (不足時はfalseを返す)
func (bp *BlackJackPlayer) SubtractChips(amount int) bool {
	if bp.chips < amount {
		return false
	}
	bp.chips -= amount
	return true
}

// IsSoft ソフトハンド（11として有効なエースを含む）かどうか判定
func (bp *BlackJackPlayer) IsSoft() bool {
	_, isSoft := calcScore(bp.cards)
	return isSoft
}
