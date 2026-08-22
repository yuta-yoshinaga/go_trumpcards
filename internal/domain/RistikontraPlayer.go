//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// RistikontraPlayer は Pişti のプレイヤー。
// 基底の GamePlayer (手札) に加えて、捕獲した札の山と、累積した Pişti ボーナスを持つ。
type RistikontraPlayer struct {
	*GamePlayer
	capturedCards []*Card
}

// NewRistikontraPlayer は RistikontraPlayer を構築する。
func NewRistikontraPlayer(isHuman bool) *RistikontraPlayer {
	return &RistikontraPlayer{
		GamePlayer:    NewGamePlayer(isHuman),
		capturedCards: make([]*Card, 0),
	}
}

// GetCapturedCards は捕獲した札を取得する。
func (p *RistikontraPlayer) GetCapturedCards() []*Card { return p.capturedCards }

// CapturedCount は捕獲した札の枚数を返す。
func (p *RistikontraPlayer) CapturedCount() int { return len(p.capturedCards) }

// AddCaptured は札を捕獲山に追加する。
func (p *RistikontraPlayer) AddCaptured(cards []*Card) {
	p.capturedCards = append(p.capturedCards, cards...)
}

// RemoveCaptured は捕獲山の末尾から cards を取り除く。
//
// **打ち返しで奪われたときに使う。** リスティコントラでは、捕獲した直後に
// 同ランクを被せられると束ごと相手に移る。直前の捕獲ぶんは必ず末尾に
// 積まれているので、末尾から同じ枚数を外せばよい。枚数が足りない盤面は
// 起こらないが、来たぶんだけ外して壊れないようにしてある。
func (p *RistikontraPlayer) RemoveCaptured(cards []*Card) {
	n := len(cards)
	if n <= 0 {
		return
	}
	if n > len(p.capturedCards) {
		n = len(p.capturedCards)
	}
	p.capturedCards = p.capturedCards[:len(p.capturedCards)-n]
}

// ResetCaptured は捕獲山をクリアする (新規ゲームの先頭で呼ぶ)。
func (p *RistikontraPlayer) ResetCaptured() {
	p.capturedCards = make([]*Card, 0)
}

// ristikontraPlayerJSON is the JSON wire format for RistikontraPlayer.
type ristikontraPlayerJSON struct {
	GamePlayer    *GamePlayer `json:"gp"`
	CapturedCards []*Card     `json:"cc"`
}

// MarshalJSON implements json.Marshaler.
func (p *RistikontraPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(ristikontraPlayerJSON{
		GamePlayer:    p.GamePlayer,
		CapturedCards: p.capturedCards,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *RistikontraPlayer) UnmarshalJSON(data []byte) error {
	var j ristikontraPlayerJSON
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
