package domain

// OldMaidPlayer ババ抜きプレイヤークラス
type OldMaidPlayer struct {
	Player
	isHuman    bool
	isFinished bool
}

// NewOldMaidPlayer コンストラクタ
func NewOldMaidPlayer(isHuman bool) *OldMaidPlayer {
	return &OldMaidPlayer{
		Player:     Player{cards: make([]*Card, 0)},
		isHuman:    isHuman,
		isFinished: false,
	}
}

// GetIsHuman 人間プレイヤーかどうか
func (p *OldMaidPlayer) GetIsHuman() bool {
	return p.isHuman
}

// GetIsFinished 上がっているかどうか
func (p *OldMaidPlayer) GetIsFinished() bool {
	return p.isFinished
}

// SetIsFinished 上がり状態設定
func (p *OldMaidPlayer) SetIsFinished(v bool) {
	p.isFinished = v
}

// DiscardPairs ペアのカードを捨てる (捨てたカードとペア数を返す)
func (p *OldMaidPlayer) DiscardPairs() ([]*Card, int) {
	discardedCards := make([]*Card, 0)
	pairs := 0
	for {
		found := false
		for i := 0; i < len(p.cards); i++ {
			c1 := p.cards[i]
			if c1.GetDesign() == CardDesignJoker {
				continue
			}
			for j := i + 1; j < len(p.cards); j++ {
				c2 := p.cards[j]
				if c2.GetDesign() == CardDesignJoker {
					continue
				}
				if c1.GetValue() == c2.GetValue() {
					newCards := make([]*Card, 0, len(p.cards)-2)
					for k, c := range p.cards {
						if k != i && k != j {
							newCards = append(newCards, c)
						}
					}
					p.cards = newCards
					discardedCards = append(discardedCards, c1, c2)
					pairs++
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			break
		}
	}
	return discardedCards, pairs
}
