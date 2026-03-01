package domain

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

// ShuffleCards 手札をランダムに並び替える。
// Go 1.20 以降はグローバル乱数生成器が起動時に自動でランダムシードされるため、
// 追加のシード設定は不要。
func (p *Player) ShuffleCards() {
	rand.Shuffle(len(p.cards), func(i, j int) {
		p.cards[i], p.cards[j] = p.cards[j], p.cards[i]
	})
}

// ReorderCards indices で指定された順番に手札を並び替える。
// indices は [0, len(cards)) の有効な順列でなければならない。
func (p *Player) ReorderCards(indices []int) error {
	n := len(p.cards)
	if len(indices) != n {
		return ErrInvalidIndices
	}
	used := make([]bool, n)
	for _, idx := range indices {
		if idx < 0 || idx >= n || used[idx] {
			return ErrInvalidIndices
		}
		used[idx] = true
	}
	newCards := make([]*Card, n)
	for i, idx := range indices {
		newCards[i] = p.cards[idx]
	}
	p.cards = newCards
	return nil
}

// PrependCard カードを手札の先頭に追加
func (p *Player) PrependCard(card *Card) {
	p.cards = append([]*Card{card}, p.cards...)
}
