package entities

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
