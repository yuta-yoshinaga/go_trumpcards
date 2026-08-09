//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// MushiPlayer は虫のプレイヤー。
type MushiPlayer struct {
	*GamePlayer
}

// NewMushiPlayer はコンストラクタ。
func NewMushiPlayer(isHuman bool) *MushiPlayer {
	return &MushiPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame は手札と上がり状態を初期化する。
func (p *MushiPlayer) ResetGame() {
	resetPlayer(p)
}

// mushiPlayerJSON is the JSON wire format for MushiPlayer.
type mushiPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *MushiPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(mushiPlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *MushiPlayer) UnmarshalJSON(data []byte) error {
	var j mushiPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	return nil
}
