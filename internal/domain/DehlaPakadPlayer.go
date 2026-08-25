//go:build !js || !wasm || extra

package domain

import "encoding/json"

// DehlaPakadPlayer はデーラ・パカドのプレイヤー。
//
// **取ったトリックは席ではなくチームに積まれる。** ここが持つ TrickHolder は
// 「この席が中央の山を引き取った回数」であって、チームの 10 の枚数はゲーム側で
// 数える。
type DehlaPakadPlayer struct {
	*GamePlayer
	*TrickHolder
}

// NewDehlaPakadPlayer constructs a DehlaPakadPlayer.
func NewDehlaPakadPlayer(isHuman bool) *DehlaPakadPlayer {
	return &DehlaPakadPlayer{
		GamePlayer:  NewGamePlayer(isHuman),
		TrickHolder: &TrickHolder{},
	}
}

// ResetDeal は 1 ハンド分の状態をクリアする。
func (p *DehlaPakadPlayer) ResetDeal() {
	p.Reset()
	p.ResetTricks()
	p.SetIsFinished(false)
}

// dehlaPakadPlayerJSON is the JSON wire format for DehlaPakadPlayer.
type dehlaPakadPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *DehlaPakadPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(dehlaPakadPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *DehlaPakadPlayer) UnmarshalJSON(data []byte) error {
	var j dehlaPakadPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.TrickHolder != nil {
		p.TrickHolder = j.TrickHolder
	} else {
		p.TrickHolder = &TrickHolder{}
	}
	return nil
}
