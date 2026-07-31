//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// BostonPlayer はボストンのプレイヤークラス。
type BostonPlayer struct {
	*GamePlayer
}

// NewBostonPlayer コンストラクタ
func NewBostonPlayer(isHuman bool) *BostonPlayer {
	return &BostonPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound は局開始時に手札を初期化する。
func (p *BostonPlayer) ResetRound() {
	p.Reset()
	p.SetIsFinished(false)
}

// bostonPlayerJSON is the JSON wire format for BostonPlayer.
type bostonPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *BostonPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(bostonPlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BostonPlayer) UnmarshalJSON(data []byte) error {
	var j bostonPlayerJSON
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
