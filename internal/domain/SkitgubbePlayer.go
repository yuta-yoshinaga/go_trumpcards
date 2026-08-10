//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// SkitgubbePlayer はシートグッベのプレイヤー。
type SkitgubbePlayer struct {
	*GamePlayer
}

// NewSkitgubbePlayer はコンストラクタ。
func NewSkitgubbePlayer(isHuman bool) *SkitgubbePlayer {
	return &SkitgubbePlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame は手札と上がり状態を初期化する。
func (p *SkitgubbePlayer) ResetGame() {
	resetPlayer(p)
}

// skitgubbePlayerJSON is the JSON wire format for SkitgubbePlayer.
type skitgubbePlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *SkitgubbePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(skitgubbePlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SkitgubbePlayer) UnmarshalJSON(data []byte) error {
	var j skitgubbePlayerJSON
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
