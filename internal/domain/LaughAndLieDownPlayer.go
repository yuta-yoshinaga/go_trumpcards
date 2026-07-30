//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// LaughAndLieDownPlayer は Laugh and Lie Down のプレイヤー。
type LaughAndLieDownPlayer struct {
	*GamePlayer
}

// NewLaughAndLieDownPlayer はコンストラクタ。
func NewLaughAndLieDownPlayer(isHuman bool) *LaughAndLieDownPlayer {
	return &LaughAndLieDownPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame は手札と上がり状態を初期化する。
func (p *LaughAndLieDownPlayer) ResetGame() {
	p.Reset()
	p.SetIsFinished(false)
}

// laughAndLieDownPlayerJSON is the JSON wire format for LaughAndLieDownPlayer.
type laughAndLieDownPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *LaughAndLieDownPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(laughAndLieDownPlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *LaughAndLieDownPlayer) UnmarshalJSON(data []byte) error {
	var raw laughAndLieDownPlayerJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.GamePlayer == nil {
		raw.GamePlayer = NewGamePlayer(false)
	}
	p.GamePlayer = raw.GamePlayer
	return nil
}
