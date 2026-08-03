//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// KaiserPlayer はカイザーのプレイヤークラス。
type KaiserPlayer struct {
	*GamePlayer
}

// NewKaiserPlayer コンストラクタ
func NewKaiserPlayer(isHuman bool) *KaiserPlayer {
	return &KaiserPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// GetTeam は所属チームを返す。
func (p *KaiserPlayer) GetTeam(seat int) int { return KaiserTeamOf(seat) }

// ResetRound は局開始時に手札を初期化する。
func (p *KaiserPlayer) ResetRound() {
	p.Reset()
	p.SetIsFinished(false)
}

// kaiserPlayerJSON is the JSON wire format for KaiserPlayer.
type kaiserPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *KaiserPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(kaiserPlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *KaiserPlayer) UnmarshalJSON(data []byte) error {
	var j kaiserPlayerJSON
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
