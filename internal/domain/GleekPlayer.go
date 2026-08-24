//go:build !js || !wasm || extra

package domain

import "encoding/json"

// GleekPlayer グリーク (Gleek) のプレイヤークラス。手札 (GamePlayer) と
// 獲得トリック (TrickHolder) を保持する。点はゲーム本体が管理する。
type GleekPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewGleekPlayer コンストラクタ
func NewGleekPlayer(isHuman bool) *GleekPlayer {
	return &GleekPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット (トリック・手札・終了状態を初期化)
func (p *GleekPlayer) ResetRound() {
	resetPlayerRound(p)
}

// gleekPlayerJSON is the JSON wire format for GleekPlayer.
type gleekPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *GleekPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(gleekPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *GleekPlayer) UnmarshalJSON(data []byte) error {
	var j gleekPlayerJSON
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
