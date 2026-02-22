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
	for i := 0; i < len(bp.cards); i++ {
		value := bp.cards[i].GetValue()
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
		tmpScore1 := score + 1
		tmpScore2 := score + 11
		if 22 <= tmpScore1 && 22 <= tmpScore2 {
			// どちらもバーストしているならエースを1
			score = tmpScore1
		} else if tmpScore1 <= 21 && 22 <= tmpScore2 {
			// エースが11でバーストしているならエースを1
			score = tmpScore1
		} else {
			// どちらもバーストしていないならエースを11
			score = tmpScore2
		}
	}
	return score
}
