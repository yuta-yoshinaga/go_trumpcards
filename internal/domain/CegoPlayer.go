//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// CegoPlayer チェゴ (Cego) のプレイヤークラス。手札 (GamePlayer) と獲得トリック (TrickHolder) を
// 保持する。得点はゲーム本体が管理する。
type CegoPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewCegoPlayer コンストラクタ
func NewCegoPlayer(isHuman bool) *CegoPlayer {
	return &CegoPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット (トリック・手札・終了状態を初期化)
func (p *CegoPlayer) ResetRound() {
	resetPlayerRound(p)
}

// cegoPlayerJSON is the JSON wire format for CegoPlayer.
type cegoPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *CegoPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(cegoPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *CegoPlayer) UnmarshalJSON(data []byte) error {
	var j cegoPlayerJSON
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
