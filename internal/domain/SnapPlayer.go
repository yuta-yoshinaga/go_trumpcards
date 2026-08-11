//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
)

// SnapPlayer はスナップのプレイヤー。
//
// 各プレイヤーは裏向きのストックを 1 つだけ持ちます。獲得した場札はストックの
// 底に戻されるので、独立した獲得置き場はありません。
type SnapPlayer struct {
	*GamePlayer
	stock []*Card
}

// NewSnapPlayer はコンストラクタ。
func NewSnapPlayer(isHuman bool) *SnapPlayer {
	return &SnapPlayer{GamePlayer: NewGamePlayer(isHuman), stock: make([]*Card, 0)}
}

// GetStockSize はストックの残枚数を返す。
func (p *SnapPlayer) GetStockSize() int { return len(p.stock) }

// HasStock はストックにカードが残っているかを返す。
func (p *SnapPlayer) HasStock() bool { return len(p.stock) > 0 }

// AddToStockBottom はストックの底にカードを追加する。
func (p *SnapPlayer) AddToStockBottom(cards ...*Card) { p.stock = append(p.stock, cards...) }

// DrawTop はストックの先頭から 1 枚引く（空なら nil）。
func (p *SnapPlayer) DrawTop() *Card {
	if len(p.stock) == 0 {
		return nil
	}
	c := p.stock[0]
	p.stock = p.stock[1:]
	return c
}

// ResetStock はストックを空にする。
func (p *SnapPlayer) ResetStock() { p.stock = make([]*Card, 0) }

// snapPlayerJSON is the JSON wire format for SnapPlayer.
type snapPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Stock      []*Card     `json:"st"`
}

// MarshalJSON implements json.Marshaler.
func (p *SnapPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(snapPlayerJSON{GamePlayer: p.GamePlayer, Stock: p.stock})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SnapPlayer) UnmarshalJSON(data []byte) error {
	var j snapPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > snapMaxSliceLen {
		return errors.New("snap: input array exceeds maximum allowed size")
	}
	// **枚数だけでなく中身も見る (#5310 の再発防止)。**
	for _, c := range j.Stock {
		if c == nil {
			return errors.New("nil card in the stock")
		}
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
