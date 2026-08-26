//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// PiedmonteseTarotPlayer はピエモンテ・タロッコの 1 席。手札 (GamePlayer) と獲得
// トリック (TrickHolder) を持つ。得点はゲーム本体が管理する。
type PiedmonteseTarotPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewPiedmonteseTarotPlayer コンストラクタ。
func NewPiedmonteseTarotPlayer(isHuman bool) *PiedmonteseTarotPlayer {
	return &PiedmonteseTarotPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット (トリック・手札・終了状態を初期化)。
func (p *PiedmonteseTarotPlayer) ResetRound() {
	resetPlayerRound(p)
}

// piedmonteseTarotPlayerJSON is the JSON wire format for PiedmonteseTarotPlayer.
type piedmonteseTarotPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *PiedmonteseTarotPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(piedmonteseTarotPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *PiedmonteseTarotPlayer) UnmarshalJSON(data []byte) error {
	var j piedmonteseTarotPlayerJSON
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
