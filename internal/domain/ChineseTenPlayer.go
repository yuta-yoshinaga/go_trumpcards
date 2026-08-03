//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// ChineseTenPlayer は撿紅點のプレイヤー。
type ChineseTenPlayer struct {
	*GamePlayer
}

// NewChineseTenPlayer はコンストラクタ。
func NewChineseTenPlayer(isHuman bool) *ChineseTenPlayer {
	return &ChineseTenPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame は手札と上がり状態を初期化する。
func (p *ChineseTenPlayer) ResetGame() {
	p.Reset()
	p.SetIsFinished(false)
}

// chineseTenPlayerJSON is the JSON wire format for ChineseTenPlayer.
type chineseTenPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *ChineseTenPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(chineseTenPlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ChineseTenPlayer) UnmarshalJSON(data []byte) error {
	var j chineseTenPlayerJSON
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
