//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// CuarentaPlayer クアレンタのプレイヤー。
// 基底の GamePlayer (手札) に加えて、自チームが捕獲したカードは
// チーム単位で集計するため、ここでは個人の捕獲札のみを保持する。
type CuarentaPlayer struct {
	*GamePlayer
	capturedCards []*Card
}

// NewCuarentaPlayer constructs a CuarentaPlayer.
func NewCuarentaPlayer(isHuman bool) *CuarentaPlayer {
	return &CuarentaPlayer{
		GamePlayer:    NewGamePlayer(isHuman),
		capturedCards: make([]*Card, 0),
	}
}

// GetCapturedCards 獲得した捕獲札を取得。
func (p *CuarentaPlayer) GetCapturedCards() []*Card { return p.capturedCards }

// CapturedCount 獲得した捕獲札の枚数。
func (p *CuarentaPlayer) CapturedCount() int { return len(p.capturedCards) }

// AddCaptured カードを捕獲札に追加。
func (p *CuarentaPlayer) AddCaptured(cards []*Card) {
	p.capturedCards = append(p.capturedCards, cards...)
}

// ResetCaptured 捕獲札をクリア (新ラウンドの先頭で呼ぶ)。
func (p *CuarentaPlayer) ResetCaptured() {
	p.capturedCards = make([]*Card, 0)
}

// cuarentaPlayerJSON is the JSON wire format for CuarentaPlayer.
type cuarentaPlayerJSON struct {
	GamePlayer    *GamePlayer `json:"gp"`
	CapturedCards []*Card     `json:"cc"`
}

// MarshalJSON implements json.Marshaler.
func (p *CuarentaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(cuarentaPlayerJSON{
		GamePlayer:    p.GamePlayer,
		CapturedCards: p.capturedCards,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *CuarentaPlayer) UnmarshalJSON(data []byte) error {
	var j cuarentaPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.capturedCards = j.CapturedCards
	if p.capturedCards == nil {
		p.capturedCards = make([]*Card, 0)
	}
	return nil
}
