package domain

import "encoding/json"

// EgyptianRatscrewPlayer エジプシャン・ラットスクリューのプレイヤー
//
// Slapjack と同じく裏向きのストック (stock) を 1 つだけ持つ。獲得した山札 (pile)
// は stock の底に裏向きで戻されるため、独立した discardPile は持たない。
type EgyptianRatscrewPlayer struct {
	*GamePlayer
	stock []*Card
}

// NewEgyptianRatscrewPlayer コンストラクタ
func NewEgyptianRatscrewPlayer(isHuman bool) *EgyptianRatscrewPlayer {
	return &EgyptianRatscrewPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		stock:      make([]*Card, 0),
	}
}

// GetStockSize ストック残枚数
func (p *EgyptianRatscrewPlayer) GetStockSize() int { return len(p.stock) }

// HasStock ストックにカードが残っているか
func (p *EgyptianRatscrewPlayer) HasStock() bool { return len(p.stock) > 0 }

// AddToStockBottom ストックの底にカードを追加する (獲得カード/ペナルティ受領用)
func (p *EgyptianRatscrewPlayer) AddToStockBottom(cards ...*Card) {
	p.stock = append(p.stock, cards...)
}

// AddToStockTop ストックの先頭にカードを追加する (主にテスト用)
func (p *EgyptianRatscrewPlayer) AddToStockTop(cards ...*Card) {
	p.stock = append(append([]*Card{}, cards...), p.stock...)
}

// DrawTop ストックの先頭から1枚引く (空なら nil)
func (p *EgyptianRatscrewPlayer) DrawTop() *Card {
	if len(p.stock) == 0 {
		return nil
	}
	c := p.stock[0]
	p.stock = p.stock[1:]
	return c
}

// ResetStock ストックを空にする
func (p *EgyptianRatscrewPlayer) ResetStock() {
	p.stock = make([]*Card, 0)
}

// egyptianRatscrewPlayerJSON is the JSON wire format for EgyptianRatscrewPlayer.
type egyptianRatscrewPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Stock      []*Card     `json:"st"`
}

// MarshalJSON implements json.Marshaler.
func (p *EgyptianRatscrewPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(egyptianRatscrewPlayerJSON{
		GamePlayer: p.GamePlayer,
		Stock:      p.stock,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *EgyptianRatscrewPlayer) UnmarshalJSON(data []byte) error {
	var j egyptianRatscrewPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.stock = j.Stock
	if p.stock == nil {
		p.stock = make([]*Card, 0)
	}
	return nil
}
