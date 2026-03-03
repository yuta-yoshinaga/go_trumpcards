package domain

// BlackJackPlayer ブラックジャックプレイヤークラス
type BlackJackPlayer struct {
	Player     // 親クラス
	ChipHolder // チップ管理
}

// NewBlackJackPlayer コンストラクタ
func NewBlackJackPlayer() *BlackJackPlayer {
	return &BlackJackPlayer{
		Player: Player{
			cards: make([]*Card, 0),
		},
	}
}

// GetScore 手札から現在のスコア計算
func (bp *BlackJackPlayer) GetScore() int {
	return CalculateBlackJackScore(bp.cards)
}

// IsSoft ソフトハンド（11として有効なエースを含む）かどうか判定
func (bp *BlackJackPlayer) IsSoft() bool {
	_, isSoft := calcScore(bp.cards)
	return isSoft
}
