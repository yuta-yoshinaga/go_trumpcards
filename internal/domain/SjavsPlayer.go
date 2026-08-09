//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// SjavsPlayer はシャウスのプレイヤー。
type SjavsPlayer struct {
	*GamePlayer
}

// NewSjavsPlayer はコンストラクタ。
func NewSjavsPlayer(isHuman bool) *SjavsPlayer {
	return &SjavsPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame は手札と上がり状態を初期化する。
func (p *SjavsPlayer) ResetGame() {
	resetPlayer(p)
}

// sjavsPlayerJSON is the JSON wire format for SjavsPlayer.
type sjavsPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *SjavsPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(sjavsPlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SjavsPlayer) UnmarshalJSON(data []byte) error {
	var raw sjavsPlayerJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.GamePlayer == nil {
		raw.GamePlayer = NewGamePlayer(false)
	}
	p.GamePlayer = raw.GamePlayer
	return nil
}
