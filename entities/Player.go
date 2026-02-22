package entities

import "math/rand"

// Player プレイヤークラス
type Player struct {
	cards []*Card // プレイヤーカード
}

// NewPlayer コンストラクタ
func NewPlayer() *Player {
	return &Player{
		cards: make([]*Card, 0),
	}
}

// GetCardsSize プレイヤーカードの枚数取得
func (p *Player) GetCardsSize() int {
	return len(p.cards)
}

// AddCard カード追加
func (p *Player) AddCard(card *Card) {
	p.cards = append(p.cards, card)
}

// GetCard 指定番目のカード取得
func (p *Player) GetCard(idx int) *Card {
	var res *Card = nil
	if 0 <= idx && idx < len(p.cards) {
		res = p.cards[idx]
	}
	return res
}

// Reset カードリセット
func (p *Player) Reset() {
	p.cards = make([]*Card, 0)
}

// ShuffleCards 手札をランダムに並び替える
func (p *Player) ShuffleCards() {
	rand.Shuffle(len(p.cards), func(i, j int) {
		p.cards[i], p.cards[j] = p.cards[j], p.cards[i]
	})
}
