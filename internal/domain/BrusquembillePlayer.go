//go:build !js || !wasm || classic

package domain

import "encoding/json"

// BrusquembillePlayer ブリュスカンビーユプレイヤークラス
type BrusquembillePlayer struct {
	*GamePlayer
	TrickHolder
}

// NewBrusquembillePlayer コンストラクタ
func NewBrusquembillePlayer(isHuman bool) *BrusquembillePlayer {
	return &BrusquembillePlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetGame ゲームをリセット (手札/トリック/上がり状態を初期化)
func (p *BrusquembillePlayer) ResetGame() {
	resetPlayerRound(p)
}

// brusquembillePlayerJSON is the JSON wire format for BrusquembillePlayer.
type brusquembillePlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *BrusquembillePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(brusquembillePlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BrusquembillePlayer) UnmarshalJSON(data []byte) error {
	var j brusquembillePlayerJSON
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
