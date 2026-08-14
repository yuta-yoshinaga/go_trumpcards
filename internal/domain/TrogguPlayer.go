//go:build !js || !wasm || extra

package domain

import "encoding/json"

// TrogguPlayer トロッグ (Troggu) のプレイヤー。手札 (GamePlayer) と獲得トリック
// (TrickHolder) を保持する。得点はゲーム本体が管理する。
type TrogguPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewTrogguPlayer コンストラクタ
func NewTrogguPlayer(isHuman bool) *TrogguPlayer {
	return &TrogguPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット (トリック・手札・終了状態を初期化)
func (p *TrogguPlayer) ResetRound() {
	resetPlayerRound(p)
}

// trogguPlayerJSON is the JSON wire format for TrogguPlayer.
type trogguPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *TrogguPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(trogguPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *TrogguPlayer) UnmarshalJSON(data []byte) error {
	var j trogguPlayerJSON
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
