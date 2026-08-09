//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// DesmochePlayer デスモチェのプレイヤークラス。
//
// メルドは場 (Desmoche.melds) 側で所有者付きで持つ。ここが Conquian などと違う
// のは、**desmoche は他人のメルドではなく自分のメルドを組み替える手**なので、
// 場に出た札を席ごとに分けて持つ必要がないためである。
type DesmochePlayer struct {
	*GamePlayer
}

// NewDesmochePlayer コンストラクタ
func NewDesmochePlayer(isHuman bool) *DesmochePlayer {
	return &DesmochePlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンド開始時に手札を空にする
func (p *DesmochePlayer) ResetRound() {
	resetPlayer(p)
}

// desmochePlayerJSON is the JSON wire format for DesmochePlayer.
type desmochePlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *DesmochePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(desmochePlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *DesmochePlayer) UnmarshalJSON(data []byte) error {
	var j desmochePlayerJSON
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
