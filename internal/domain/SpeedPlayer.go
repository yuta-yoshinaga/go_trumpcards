package domain

import "encoding/json"

// SpeedPlayer スピードプレイヤークラス
type SpeedPlayer struct {
	*GamePlayer
	drawPile []*Card // 山札
}

// NewSpeedPlayer コンストラクタ
func NewSpeedPlayer(isHuman bool) *SpeedPlayer {
	return &SpeedPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		drawPile:   make([]*Card, 0),
	}
}

// GetDrawPileSize 山札の枚数を返す
func (p *SpeedPlayer) GetDrawPileSize() int {
	return len(p.drawPile)
}

// AddToDrawPile 山札にカードを追加する
func (p *SpeedPlayer) AddToDrawPile(cards ...*Card) {
	p.drawPile = append(p.drawPile, cards...)
}

// DrawToHand 山札から1枚引いて手札に加える。山札が空なら false を返す。
func (p *SpeedPlayer) DrawToHand() bool {
	if len(p.drawPile) == 0 {
		return false
	}
	card := p.drawPile[0]
	p.drawPile = p.drawPile[1:]
	p.AddCard(card)
	return true
}

// RefillHand 手札が maxSize になるまで山札から引く
func (p *SpeedPlayer) RefillHand(maxSize int) {
	for p.GetCardsSize() < maxSize && len(p.drawPile) > 0 {
		p.DrawToHand()
	}
}

// HasCards 手札または山札にカードが残っているか
func (p *SpeedPlayer) HasCards() bool {
	return p.GetCardsSize() > 0 || len(p.drawPile) > 0
}

// ResetDrawPile 山札をリセットする
func (p *SpeedPlayer) ResetDrawPile() {
	p.drawPile = make([]*Card, 0)
}

// speedPlayerJSON is the JSON wire format for SpeedPlayer.
type speedPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	DrawPile   []*Card     `json:"dp"`
}

// MarshalJSON implements json.Marshaler.
func (p *SpeedPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(speedPlayerJSON{
		GamePlayer: p.GamePlayer,
		DrawPile:   p.drawPile,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SpeedPlayer) UnmarshalJSON(data []byte) error {
	var j speedPlayerJSON
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
	return nil
}
