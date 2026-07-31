//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// LobaPlayer はロバのプレイヤー。
type LobaPlayer struct {
	*GamePlayer
}

// NewLobaPlayer はコンストラクタ。
func NewLobaPlayer(isHuman bool) *LobaPlayer {
	return &LobaPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame は手札と上がり状態を初期化する。
func (p *LobaPlayer) ResetGame() {
	p.Reset()
	p.SetIsFinished(false)
}

// lobaPlayerJSON is the JSON wire format for LobaPlayer.
type lobaPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *LobaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(lobaPlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *LobaPlayer) UnmarshalJSON(data []byte) error {
	var raw lobaPlayerJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.GamePlayer == nil {
		raw.GamePlayer = NewGamePlayer(false)
	}
	p.GamePlayer = raw.GamePlayer
	return nil
}
