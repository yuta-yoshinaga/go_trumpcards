//go:build !js || !wasm || classic

package domain

import "encoding/json"

// ShengJiPlayer は升级のプレイヤークラス。
type ShengJiPlayer struct {
	*GamePlayer
}

// NewShengJiPlayer コンストラクタ
func NewShengJiPlayer(isHuman bool) *ShengJiPlayer {
	return &ShengJiPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// GetTeam は席の所属チームを返す。
func (p *ShengJiPlayer) GetTeam(seat int) int { return ShengJiTeamOf(seat) }

// ResetRound は局開始時に手札を初期化する。
func (p *ShengJiPlayer) ResetRound() {
	resetPlayer(p)
}

// shengJiPlayerJSON is the JSON wire format for ShengJiPlayer.
type shengJiPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *ShengJiPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(shengJiPlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ShengJiPlayer) UnmarshalJSON(data []byte) error {
	var j shengJiPlayerJSON
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
