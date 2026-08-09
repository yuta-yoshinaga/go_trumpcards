//go:build !js || !wasm || solo

package domain

import "encoding/json"

// SchnapsenPlayer シュナプセンプレイヤークラス
type SchnapsenPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewSchnapsenPlayer コンストラクタ
func NewSchnapsenPlayer(isHuman bool) *SchnapsenPlayer {
	return &SchnapsenPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetGame ゲームをリセット (手札/トリック/上がり状態を初期化)
func (p *SchnapsenPlayer) ResetGame() {
	resetPlayerRound(p)
}

// schnapsenPlayerJSON is the JSON wire format for SchnapsenPlayer.
type schnapsenPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *SchnapsenPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(schnapsenPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SchnapsenPlayer) UnmarshalJSON(data []byte) error {
	var j schnapsenPlayerJSON
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
