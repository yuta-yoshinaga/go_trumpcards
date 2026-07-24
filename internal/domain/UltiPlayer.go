//go:build !js || !wasm || solo

package domain

import "encoding/json"

// UltiPlayer ウルティ (Ulti) のプレイヤークラス。手札 (GamePlayer) と
// 獲得トリック (TrickHolder) を保持する。コインはゲーム本体が管理する。
type UltiPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewUltiPlayer コンストラクタ
func NewUltiPlayer(isHuman bool) *UltiPlayer {
	return &UltiPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット (トリック・手札・終了状態を初期化)
func (p *UltiPlayer) ResetRound() {
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
}

// ultiPlayerJSON is the JSON wire format for UltiPlayer.
type ultiPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *UltiPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(ultiPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *UltiPlayer) UnmarshalJSON(data []byte) error {
	var j ultiPlayerJSON
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
