//go:build !js || !wasm || classic

package domain

import "encoding/json"

// UnsunKarutaPlayer はうんすんカルタの 1 席。手札と獲得トリックを持つ。
// 「コ」(トリック) の集計はチーム単位なのでゲーム本体が持つ。
type UnsunKarutaPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewUnsunKarutaPlayer コンストラクタ。
func NewUnsunKarutaPlayer(isHuman bool) *UnsunKarutaPlayer {
	return &UnsunKarutaPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound はラウンドをリセットする。
func (p *UnsunKarutaPlayer) ResetRound() {
	resetPlayerRound(p)
}

// unsunKarutaPlayerJSON is the JSON wire format for UnsunKarutaPlayer.
type unsunKarutaPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *UnsunKarutaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(unsunKarutaPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *UnsunKarutaPlayer) UnmarshalJSON(data []byte) error {
	var j unsunKarutaPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	return nil
}
