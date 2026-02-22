package entities

// BlackJackPlayer ブラックジャックプレイヤークラス
type BlackJackPlayer struct {
	Player     // 親クラス
	score  int // スコア
}

// NewBlackJackPlayer コンストラクタ
func NewBlackJackPlayer() *BlackJackPlayer {
	return &BlackJackPlayer{
		Player: Player{
			cards: make([]*Card, 0),
		},
		score: 0,
	}
}

// AddCard カード追加
func (bp *BlackJackPlayer) AddCard(card *Card) {
	bp.cards = append(bp.cards, card)
	bp.CalcScore()
}

// CalcScore 手札から現在のスコア計算
func (bp *BlackJackPlayer) CalcScore() {
	aceFlag := false
	bp.score = 0
	for i := 0; i < len(bp.cards); i++ {
		value := bp.cards[i].GetValue()
		if 2 <= value && value <= 10 {
			// 2～10
			bp.score += value
		} else if 11 <= value && value <= 13 {
			// 11～13
			bp.score += 10
		} else {
			if aceFlag {
				// 2枚目のエースは強制的に1で換算する
				bp.score++
			} else {
				// エースは後ほど計算する
				aceFlag = true
			}
		}
	}
	if aceFlag {
		// エース計算
		tmpScore1 := bp.score + 1
		tmpScore2 := bp.score + 11
		if 22 <= tmpScore1 && 22 <= tmpScore2 {
			// どちらもバーストしているならエースを1
			bp.score = tmpScore1
		} else if tmpScore1 <= 21 && 22 <= tmpScore2 {
			// エースが11でバーストしているならエースを1
			bp.score = tmpScore1
		} else {
			// どちらもバーストしていないならエースを11
			bp.score = tmpScore2
		}
	}
}

// GetScore スコア取得
func (bp *BlackJackPlayer) GetScore() int {
	return bp.score
}
