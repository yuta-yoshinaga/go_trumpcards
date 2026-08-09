//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// TrexPlayer はトリックスのプレイヤー。
type TrexPlayer struct {
	*GamePlayer
}

// NewTrexPlayer はコンストラクタ。
func NewTrexPlayer(isHuman bool) *TrexPlayer {
	return &TrexPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame は手札と上がり状態を初期化する。
func (p *TrexPlayer) ResetGame() {
	resetPlayer(p)
}

// trexPlayerJSON is the JSON wire format for TrexPlayer.
type trexPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *TrexPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(trexPlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *TrexPlayer) UnmarshalJSON(data []byte) error {
	var raw trexPlayerJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.GamePlayer == nil {
		raw.GamePlayer = NewGamePlayer(false)
	}
	p.GamePlayer = raw.GamePlayer
	return nil
}
