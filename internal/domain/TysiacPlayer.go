//go:build !js || !wasm || extra

package domain

import "encoding/json"

// TysiacPlayer サウザンド (Tysiąc) のプレイヤークラス。手札 (GamePlayer) と獲得トリック
// (TrickHolder) を保持する。点はゲーム本体が管理する。
type TysiacPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewTysiacPlayer コンストラクタ
func NewTysiacPlayer(isHuman bool) *TysiacPlayer {
	return &TysiacPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット (トリック・手札・終了状態を初期化)
func (p *TysiacPlayer) ResetRound() {
	resetPlayerRound(p)
}

// tysiacPlayerJSON is the JSON wire format for TysiacPlayer.
type tysiacPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *TysiacPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(tysiacPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *TysiacPlayer) UnmarshalJSON(data []byte) error {
	var j tysiacPlayerJSON
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
