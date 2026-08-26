//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// OmbrePlayer オンブル (Ombre) のプレイヤークラス。手札 (GamePlayer) と
// 獲得トリック (TrickHolder) を保持する。点はゲーム本体が管理する。
type OmbrePlayer struct {
	*GamePlayer
	TrickHolder
}

// NewOmbrePlayer コンストラクタ
func NewOmbrePlayer(isHuman bool) *OmbrePlayer {
	return &OmbrePlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット (トリック・手札・終了状態を初期化)
func (p *OmbrePlayer) ResetRound() {
	resetPlayerRound(p)
}

// ombrePlayerJSON is the JSON wire format for OmbrePlayer.
type ombrePlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *OmbrePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(ombrePlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *OmbrePlayer) UnmarshalJSON(data []byte) error {
	var j ombrePlayerJSON
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
