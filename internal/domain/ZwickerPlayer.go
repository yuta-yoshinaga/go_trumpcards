//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// ZwickerPlayer はツヴィッカーのプレイヤークラス。
//
// 手札のほかに、このディールで捕獲した札と Zwick の回数を持つ。**Zwick は
// 「場を空にした」ボーナスの回数**であって、複数組同時取りの回数ではない。
type ZwickerPlayer struct {
	*GamePlayer
	captured []*Card
	zwicks   int
}

// NewZwickerPlayer コンストラクタ
func NewZwickerPlayer(isHuman bool) *ZwickerPlayer {
	return &ZwickerPlayer{GamePlayer: NewGamePlayer(isHuman), captured: make([]*Card, 0)}
}

// GetCaptured はこのディールで捕獲した札を返す。
func (p *ZwickerPlayer) GetCaptured() []*Card { return p.captured }

// AddCaptured は捕獲した札を積む。
func (p *ZwickerPlayer) AddCaptured(cards []*Card) {
	for _, c := range cards {
		if c != nil {
			p.captured = append(p.captured, c)
		}
	}
}

// GetZwicks は Zwick の回数を返す。
func (p *ZwickerPlayer) GetZwicks() int { return p.zwicks }

// AddZwick は Zwick を 1 回加える。
func (p *ZwickerPlayer) AddZwick() { p.zwicks++ }

// ResetRound はディール開始時に手札・捕獲札・Zwick を初期化する。
func (p *ZwickerPlayer) ResetRound() {
	p.Reset()
	p.SetIsFinished(false)
	p.captured = make([]*Card, 0)
	p.zwicks = 0
}

// zwickerPlayerJSON is the JSON wire format for ZwickerPlayer.
type zwickerPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Captured   []*Card     `json:"cp"`
	Zwicks     int         `json:"zw"`
}

// MarshalJSON implements json.Marshaler.
func (p *ZwickerPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(zwickerPlayerJSON{
		GamePlayer: p.GamePlayer, Captured: p.captured, Zwicks: p.zwicks,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ZwickerPlayer) UnmarshalJSON(data []byte) error {
	var j zwickerPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.captured = j.Captured
	if p.captured == nil {
		p.captured = make([]*Card, 0)
	}
	p.zwicks = j.Zwicks
	if p.zwicks < 0 {
		p.zwicks = 0
	}
	return nil
}
