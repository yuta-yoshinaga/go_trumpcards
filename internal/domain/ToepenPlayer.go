//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// ToepenPlayer はトゥーペンのプレイヤー。
type ToepenPlayer struct {
	*GamePlayer
}

// NewToepenPlayer はコンストラクタ。
func NewToepenPlayer(isHuman bool) *ToepenPlayer {
	return &ToepenPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame は手札と上がり状態を初期化する。
func (p *ToepenPlayer) ResetGame() {
	p.Reset()
	p.SetIsFinished(false)
}

// toepenPlayerJSON is the JSON wire format for ToepenPlayer.
type toepenPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *ToepenPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(toepenPlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ToepenPlayer) UnmarshalJSON(data []byte) error {
	var j toepenPlayerJSON
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
