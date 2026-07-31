//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// BidEuchrePlayer はビッド・ユーカーのプレイヤークラス。
type BidEuchrePlayer struct {
	*GamePlayer
}

// NewBidEuchrePlayer コンストラクタ
func NewBidEuchrePlayer(isHuman bool) *BidEuchrePlayer {
	return &BidEuchrePlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// GetTeam は所属チームを返す。
func (p *BidEuchrePlayer) GetTeam(seat int) int { return BidEuchreTeamOf(seat) }

// ResetRound は局開始時に手札を初期化する。
func (p *BidEuchrePlayer) ResetRound() {
	p.Reset()
	p.SetIsFinished(false)
}

// bidEuchrePlayerJSON is the JSON wire format for BidEuchrePlayer.
type bidEuchrePlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *BidEuchrePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(bidEuchrePlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BidEuchrePlayer) UnmarshalJSON(data []byte) error {
	var j bidEuchrePlayerJSON
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
