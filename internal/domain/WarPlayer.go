package domain

import (
	"encoding/json"
	"math/rand"
)

// WarPlayer 戦争ゲームのプレイヤー
//
// 手札 (base Player.cards) は使用せず、伏せた山札 (drawPile) と
// 獲得カードを入れる捨て札 (discardPile) の2つを管理する。
type WarPlayer struct {
	*GamePlayer
	drawPile    []*Card
	discardPile []*Card
}

// NewWarPlayer コンストラクタ
func NewWarPlayer(isHuman bool) *WarPlayer {
	return &WarPlayer{
		GamePlayer:  NewGamePlayer(isHuman),
		drawPile:    make([]*Card, 0),
		discardPile: make([]*Card, 0),
	}
}

// GetDrawPileSize 山札の枚数
func (p *WarPlayer) GetDrawPileSize() int { return len(p.drawPile) }

// GetDiscardPileSize 捨て札の枚数
func (p *WarPlayer) GetDiscardPileSize() int { return len(p.discardPile) }

// TotalCards 山札 + 捨て札の合計枚数
func (p *WarPlayer) TotalCards() int {
	return len(p.drawPile) + len(p.discardPile)
}

// HasCards カードが1枚以上残っているか
func (p *WarPlayer) HasCards() bool {
	return p.TotalCards() > 0
}

// AddToDrawPile 山札の末尾にカードを追加する
func (p *WarPlayer) AddToDrawPile(cards ...*Card) {
	p.drawPile = append(p.drawPile, cards...)
}

// AddToDiscardPile 捨て札の末尾にカードを追加する
func (p *WarPlayer) AddToDiscardPile(cards ...*Card) {
	p.discardPile = append(p.discardPile, cards...)
}

// DrawOne 山札の先頭から1枚引く。山札が空の場合は
// 捨て札をシャッフルして山札に戻してから引く。どちらも空なら nil。
func (p *WarPlayer) DrawOne() *Card {
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
func (p *WarPlayer) refillDrawFromDiscard() {
	rand.Shuffle(len(p.discardPile), func(i, j int) {
		p.discardPile[i], p.discardPile[j] = p.discardPile[j], p.discardPile[i]
	})
	p.drawPile = append(p.drawPile, p.discardPile...)
	p.discardPile = p.discardPile[:0]
}

// ResetPiles 両方のパイルを空にする
func (p *WarPlayer) ResetPiles() {
	p.drawPile = make([]*Card, 0)
	p.discardPile = make([]*Card, 0)
}

// warPlayerJSON is the JSON wire format for WarPlayer.
type warPlayerJSON struct {
	GamePlayer  *GamePlayer `json:"gp"`
	DrawPile    []*Card     `json:"dp"`
	DiscardPile []*Card     `json:"cp"`
}

// MarshalJSON implements json.Marshaler.
func (p *WarPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(warPlayerJSON{
		GamePlayer:  p.GamePlayer,
		DrawPile:    p.drawPile,
		DiscardPile: p.discardPile,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *WarPlayer) UnmarshalJSON(data []byte) error {
	var j warPlayerJSON
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
