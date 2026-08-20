//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// SixBidSoloPlayer はシックスビッド・ソロのプレイヤークラス。
type SixBidSoloPlayer struct {
	*GamePlayer
}

// NewSixBidSoloPlayer コンストラクタ
func NewSixBidSoloPlayer(isHuman bool) *SixBidSoloPlayer {
	return &SixBidSoloPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound は局開始時に手札を初期化する。
func (p *SixBidSoloPlayer) ResetRound() {
	resetPlayer(p)
}

// sixBidSoloPlayerJSON is the JSON wire format for SixBidSoloPlayer.
type sixBidSoloPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *SixBidSoloPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(sixBidSoloPlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SixBidSoloPlayer) UnmarshalJSON(data []byte) error {
	var j sixBidSoloPlayerJSON
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
