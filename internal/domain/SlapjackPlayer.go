package domain

import "encoding/json"

// SlapjackPlayer スラップジャックのプレイヤー
//
// 各プレイヤーは裏向きのストック (stock) を 1 つだけ持つ。獲得した山札 (pile)
// は stock の底に裏向きで戻されるため、独立した discardPile は持たない。
type SlapjackPlayer struct {
	*GamePlayer
	stock []*Card
}

// NewSlapjackPlayer コンストラクタ
func NewSlapjackPlayer(isHuman bool) *SlapjackPlayer {
	return &SlapjackPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		stock:      make([]*Card, 0),
	}
}

// GetStockSize ストック残枚数
func (p *SlapjackPlayer) GetStockSize() int { return len(p.stock) }

// HasStock ストックにカードが残っているか
func (p *SlapjackPlayer) HasStock() bool { return len(p.stock) > 0 }

// AddToStockBottom ストックの底にカードを追加する (獲得カード/ペナルティ受領用)
func (p *SlapjackPlayer) AddToStockBottom(cards ...*Card) {
	p.stock = append(p.stock, cards...)
}

// AddToStockTop ストックの先頭にカードを追加する (主にテスト用)
func (p *SlapjackPlayer) AddToStockTop(cards ...*Card) {
	p.stock = append(append([]*Card{}, cards...), p.stock...)
}

// DrawTop ストックの先頭から1枚引く (空なら nil)
func (p *SlapjackPlayer) DrawTop() *Card {
	if len(p.stock) == 0 {
		return nil
	}
	c := p.stock[0]
	p.stock = p.stock[1:]
	return c
}

// ResetStock ストックを空にする
func (p *SlapjackPlayer) ResetStock() {
	p.stock = make([]*Card, 0)
}

// slapjackPlayerJSON is the JSON wire format for SlapjackPlayer.
type slapjackPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Stock      []*Card     `json:"st"`
}

// MarshalJSON implements json.Marshaler.
func (p *SlapjackPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(slapjackPlayerJSON{
		GamePlayer: p.GamePlayer,
		Stock:      p.stock,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SlapjackPlayer) UnmarshalJSON(data []byte) error {
	var j slapjackPlayerJSON
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
