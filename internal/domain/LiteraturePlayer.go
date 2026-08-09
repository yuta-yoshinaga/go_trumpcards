//go:build !js || !wasm || solo

package domain

import "encoding/json"

// LiteraturePlayer はリテラチャーのプレイヤークラス。
type LiteraturePlayer struct {
	*GamePlayer
}

// NewLiteraturePlayer コンストラクタ
func NewLiteraturePlayer(isHuman bool) *LiteraturePlayer {
	return &LiteraturePlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// GetTeam は席の所属チームを返す。
func (p *LiteraturePlayer) GetTeam(seat int) int { return LiteratureTeamOf(seat) }

// ResetRound は局開始時に手札を初期化する。
func (p *LiteraturePlayer) ResetRound() {
	resetPlayer(p)
}

// literaturePlayerJSON is the JSON wire format for LiteraturePlayer.
type literaturePlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *LiteraturePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(literaturePlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *LiteraturePlayer) UnmarshalJSON(data []byte) error {
	var j literaturePlayerJSON
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
