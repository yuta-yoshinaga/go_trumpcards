package domain

import (
	"encoding/json"
	"math/rand"
	"sort"
)

// Player プレイヤークラス
type Player struct {
	cards []*Card // プレイヤーカード
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

// InsertCard 指定位置にカードを挿入する。位置が範囲外の場合は末尾に追加。
func (p *Player) InsertCard(card *Card, pos int) {
	if pos <= 0 {
		p.cards = append([]*Card{card}, p.cards...)
		return
	}
	if pos >= len(p.cards) {
		p.cards = append(p.cards, card)
		return
	}
	p.cards = append(p.cards[:pos], append([]*Card{card}, p.cards[pos:]...)...)
}

// RemoveCard 指定インデックスのカードを手札から取り除いて返す。
// 範囲外なら nil を返す。
func (p *Player) RemoveCard(idx int) *Card {
	if idx < 0 || idx >= len(p.cards) {
		return nil
	}
	card := p.cards[idx]
	p.cards = append(p.cards[:idx], p.cards[idx+1:]...)
	return card
}

// RemoveCards 複数インデックスのカードを手札から取り除いて返す。
// 重複インデックスは自動排除、範囲外は無視。返却順はインデックス昇順。
func (p *Player) RemoveCards(indices []int) []*Card {
	if len(indices) == 0 {
		return []*Card{}
	}
	sorted := make([]int, len(indices))
	copy(sorted, indices)
	sort.Ints(sorted)
	// deduplicate
	unique := make([]int, 0, len(sorted))
	for i, idx := range sorted {
		if i == 0 || idx != sorted[i-1] {
			unique = append(unique, idx)
		}
	}
	// collect cards (ascending order) and validate bounds
	removed := make([]*Card, 0, len(unique))
	for _, idx := range unique {
		if idx >= 0 && idx < len(p.cards) {
			removed = append(removed, p.cards[idx])
		}
	}
	// delete back-to-front to avoid index shifting
	for i := len(unique) - 1; i >= 0; i-- {
		idx := unique[i]
		if idx >= 0 && idx < len(p.cards) {
			p.cards = append(p.cards[:idx], p.cards[idx+1:]...)
		}
	}
	return removed
}

// playerJSON is the JSON wire format for Player.
type playerJSON struct {
	Cards []*Card `json:"c"`
}

// MarshalJSON implements json.Marshaler.
func (p *Player) MarshalJSON() ([]byte, error) {
	return json.Marshal(playerJSON{Cards: p.cards})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *Player) UnmarshalJSON(data []byte) error {
	var j playerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.cards = j.Cards
	if p.cards == nil {
		p.cards = make([]*Card, 0)
	}
	return nil
}
