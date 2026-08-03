//go:build !js || !wasm || extra

package domain

import "encoding/json"

// GanjifaPlayer ガンジファのプレイヤークラス。手札（GamePlayer）と獲得
// トリック（TrickHolder）を保持する。点はゲーム本体が管理する。
type GanjifaPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewGanjifaPlayer コンストラクタ
func NewGanjifaPlayer(isHuman bool) *GanjifaPlayer {
	return &GanjifaPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *GanjifaPlayer) ResetRound() {
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
}

// ganjifaPlayerJSON is the JSON wire format for GanjifaPlayer.
type ganjifaPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *GanjifaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(ganjifaPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *GanjifaPlayer) UnmarshalJSON(data []byte) error {
	var j ganjifaPlayerJSON
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
