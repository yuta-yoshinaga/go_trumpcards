package domain

import (
	"encoding/json"
	"math/rand"
)

// BeggarMyNeighbourPlayer Beggar-My-Neighbour ゲームのプレイヤー
//
// 手札 (base Player.cards) は使用せず、伏せた山札 (drawPile) と
// 獲得カードを入れる捨て札 (discardPile) の2つを管理する。
type BeggarMyNeighbourPlayer struct {
	*GamePlayer
	drawPile    []*Card
	discardPile []*Card
}

// NewBeggarMyNeighbourPlayer コンストラクタ
func NewBeggarMyNeighbourPlayer(isHuman bool) *BeggarMyNeighbourPlayer {
	return &BeggarMyNeighbourPlayer{
		GamePlayer:  NewGamePlayer(isHuman),
		drawPile:    make([]*Card, 0),
		discardPile: make([]*Card, 0),
	}
}

// GetDrawPileSize 山札の枚数
func (p *BeggarMyNeighbourPlayer) GetDrawPileSize() int { return len(p.drawPile) }

// GetDiscardPileSize 捨て札の枚数
func (p *BeggarMyNeighbourPlayer) GetDiscardPileSize() int { return len(p.discardPile) }

// TotalCards 山札 + 捨て札の合計枚数
func (p *BeggarMyNeighbourPlayer) TotalCards() int {
	return len(p.drawPile) + len(p.discardPile)
}

// HasCards カードが1枚以上残っているか
func (p *BeggarMyNeighbourPlayer) HasCards() bool {
	return p.TotalCards() > 0
}

// AddToDrawPile 山札の末尾にカードを追加する
func (p *BeggarMyNeighbourPlayer) AddToDrawPile(cards ...*Card) {
	p.drawPile = append(p.drawPile, cards...)
}

// AddToDiscardPile 捨て札の末尾にカードを追加する
func (p *BeggarMyNeighbourPlayer) AddToDiscardPile(cards ...*Card) {
	p.discardPile = append(p.discardPile, cards...)
}

// DrawOne 山札の先頭から1枚引く。山札が空の場合は
// 捨て札をシャッフルして山札に戻してから引く。どちらも空なら nil。
func (p *BeggarMyNeighbourPlayer) DrawOne() *Card {
	if len(p.drawPile) == 0 {
		if len(p.discardPile) == 0 {
			return nil
		}
		p.refillDrawFromDiscard()
	}
	card := p.drawPile[0]
	p.drawPile = p.drawPile[1:]
	return card
}

// refillDrawFromDiscard 捨て札をシャッフルして山札として戻す
func (p *BeggarMyNeighbourPlayer) refillDrawFromDiscard() {
	rand.Shuffle(len(p.discardPile), func(i, j int) {
		p.discardPile[i], p.discardPile[j] = p.discardPile[j], p.discardPile[i]
	})
	p.drawPile = append(p.drawPile, p.discardPile...)
	p.discardPile = p.discardPile[:0]
}

// ResetPiles 両方のパイルを空にする
func (p *BeggarMyNeighbourPlayer) ResetPiles() {
	p.drawPile = make([]*Card, 0)
	p.discardPile = make([]*Card, 0)
}

// beggarMyNeighbourPlayerJSON is the JSON wire format for BeggarMyNeighbourPlayer.
type beggarMyNeighbourPlayerJSON struct {
	GamePlayer  *GamePlayer `json:"gp"`
	DrawPile    []*Card     `json:"dp"`
	DiscardPile []*Card     `json:"dcp"`
}

// MarshalJSON implements json.Marshaler.
func (p *BeggarMyNeighbourPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(beggarMyNeighbourPlayerJSON{
		GamePlayer:  p.GamePlayer,
		DrawPile:    p.drawPile,
		DiscardPile: p.discardPile,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BeggarMyNeighbourPlayer) UnmarshalJSON(data []byte) error {
	var j beggarMyNeighbourPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.drawPile = j.DrawPile
	if p.drawPile == nil {
		p.drawPile = make([]*Card, 0)
	}
	p.discardPile = j.DiscardPile
	if p.discardPile == nil {
		p.discardPile = make([]*Card, 0)
	}
	return nil
}
