//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// KlaberjassPlayer はクラバーヤスのプレイヤークラス。
type KlaberjassPlayer struct {
	*GamePlayer
}

// NewKlaberjassPlayer コンストラクタ
func NewKlaberjassPlayer(isHuman bool) *KlaberjassPlayer {
	return &KlaberjassPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound はディール開始時に手札を初期化する。
func (p *KlaberjassPlayer) ResetRound() {
	resetPlayer(p)
}

// klaberjassPlayerJSON is the JSON wire format for KlaberjassPlayer.
type klaberjassPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *KlaberjassPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(klaberjassPlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *KlaberjassPlayer) UnmarshalJSON(data []byte) error {
	var j klaberjassPlayerJSON
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
