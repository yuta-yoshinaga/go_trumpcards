package entities

// BlackJackPlayer ブラックジャックプレイヤークラス
type BlackJackPlayer struct {
	Player // 親クラス
}

// NewBlackJackPlayer コンストラクタ
func NewBlackJackPlayer() *BlackJackPlayer {
	return &BlackJackPlayer{
		Player: Player{
			cards: make([]*Card, 0),
		},
	}
}

// AddCard カード追加
func (bp *BlackJackPlayer) AddCard(card *Card) {
	bp.cards = append(bp.cards, card)
}

// GetScore 手札から現在のスコア計算
func (bp *BlackJackPlayer) GetScore() int {
	aceFlag := false
	score := 0
	for _, card := range bp.cards {
		value := card.GetValue()
		if 2 <= value && value <= 10 {
			// 2～10
			score += value
		} else if 11 <= value && value <= 13 {
			// 11～13
			score += 10
		} else {
			if aceFlag {
				// 2枚目のエースは強制的に1で換算する
				score++
			} else {
				// エースは後ほど計算する
				aceFlag = true
			}
		}
	}
	if aceFlag {
		// エース計算
		if score+11 <= 21 {
			score += 11
		} else {
			score++
		}
	}
	return score
}
