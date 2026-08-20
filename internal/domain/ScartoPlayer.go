//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// ScartoPlayer スカルト (Scarto) のプレイヤークラス。手札 (GamePlayer) と獲得トリック
// (TrickHolder) を保持する。得点はゲーム本体が管理する。
type ScartoPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewScartoPlayer コンストラクタ
func NewScartoPlayer(isHuman bool) *ScartoPlayer {
	return &ScartoPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット (トリック・手札・終了状態を初期化)
func (p *ScartoPlayer) ResetRound() {
	resetPlayerRound(p)
}

// scartoPlayerJSON is the JSON wire format for ScartoPlayer.
type scartoPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *ScartoPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(scartoPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ScartoPlayer) UnmarshalJSON(data []byte) error {
	var j scartoPlayerJSON
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
