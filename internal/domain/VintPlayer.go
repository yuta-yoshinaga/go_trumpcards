//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// VintPlayer はヴィントのプレイヤークラス。
type VintPlayer struct {
	*GamePlayer
}

// NewVintPlayer コンストラクタ
func NewVintPlayer(isHuman bool) *VintPlayer {
	return &VintPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// GetTeam は所属チームを返す。
func (p *VintPlayer) GetTeam(seat int) int { return VintTeamOf(seat) }

// ResetRound は局開始時に手札を初期化する。
func (p *VintPlayer) ResetRound() {
	resetPlayer(p)
}

// vintPlayerJSON is the JSON wire format for VintPlayer.
type vintPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *VintPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(vintPlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *VintPlayer) UnmarshalJSON(data []byte) error {
	var j vintPlayerJSON
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
