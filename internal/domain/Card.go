package domain

import "encoding/json"

// カード定数
const (
	CardDesignJoker   = 0
	CardDesignSpade   = 1
	CardDesignClover  = 2
	CardDesignHeart   = 3
	CardDesignDiamond = 4
	CardDesignMin     = CardDesignJoker
	CardDesignMax     = CardDesignDiamond
	CardValueJoker    = 0
	CardValueMin      = 0
	CardValueMax      = 13
	CardCnt           = (CardValueMax * 4)
)

// Card カードクラス
type Card struct {
	design int  // カード種類
	value  int  // カード値
	draw   bool // カード払い出しフラグ
}

// NewCard コンストラクタ
func NewCard(design int, value int, draw bool) *Card {
	return &Card{
		design: design,
		value:  value,
		draw:   draw,
	}
}

// GetDesign カード種類取得
func (c *Card) GetDesign() int {
	return c.design
}

// GetValue カード値取得
func (c *Card) GetValue() int {
	return c.value
}

// SetDraw カード払い出しフラグ設定
func (c *Card) SetDraw(draw bool) {
	c.draw = draw
}

// GetDraw カード払い出しフラグ取得
func (c *Card) GetDraw() bool {
	return c.draw
}

// cardJSON is the JSON wire format for Card.
type cardJSON struct {
	D int  `json:"d"` // design
	V int  `json:"v"` // value
	W bool `json:"w"` // draw (dealt)
}

// MarshalJSON implements json.Marshaler.
func (c *Card) MarshalJSON() ([]byte, error) {
	return json.Marshal(cardJSON{D: c.design, V: c.value, W: c.draw})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Card) UnmarshalJSON(data []byte) error {
	var j cardJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.design = j.D
	c.value = j.V
	c.draw = j.W
	return nil
}
