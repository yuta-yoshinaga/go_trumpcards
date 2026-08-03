//go:build !js || !wasm || classic

package domain

import "encoding/json"

// KarnoffelPlayer はカルニッフェルのプレイヤークラス。
type KarnoffelPlayer struct {
	*GamePlayer
}

// NewKarnoffelPlayer コンストラクタ
func NewKarnoffelPlayer(isHuman bool) *KarnoffelPlayer {
	return &KarnoffelPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// GetTeam は席の所属チームを返す。
func (p *KarnoffelPlayer) GetTeam(seat int) int { return KarnoffelTeamOf(seat) }

// ResetRound は局開始時に手札を初期化する。
func (p *KarnoffelPlayer) ResetRound() {
	p.Reset()
	p.SetIsFinished(false)
}

// karnoffelPlayerJSON is the JSON wire format for KarnoffelPlayer.
type karnoffelPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *KarnoffelPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(karnoffelPlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *KarnoffelPlayer) UnmarshalJSON(data []byte) error {
	var j karnoffelPlayerJSON
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
