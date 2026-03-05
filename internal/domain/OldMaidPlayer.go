package domain

// OldMaidPlayer ババ抜きプレイヤークラス
type OldMaidPlayer struct {
	*GamePlayer
}

// NewOldMaidPlayer コンストラクタ
func NewOldMaidPlayer(isHuman bool) *OldMaidPlayer {
	return &OldMaidPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
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
