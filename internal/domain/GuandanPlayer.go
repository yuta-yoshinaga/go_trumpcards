//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// GuandanPlayer は掼蛋のプレイヤークラス。
type GuandanPlayer struct {
	*GamePlayer
}

// NewGuandanPlayer コンストラクタ
func NewGuandanPlayer(isHuman bool) *GuandanPlayer {
	return &GuandanPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// GetTeam は席の所属チームを返す。
func (p *GuandanPlayer) GetTeam(seat int) int { return GuandanTeamOf(seat) }

// ResetRound は局開始時に手札を初期化する。
func (p *GuandanPlayer) ResetRound() {
	resetPlayer(p)
}

// guandanPlayerJSON is the JSON wire format for GuandanPlayer.
type guandanPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *GuandanPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(guandanPlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *GuandanPlayer) UnmarshalJSON(data []byte) error {
	var j guandanPlayerJSON
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
