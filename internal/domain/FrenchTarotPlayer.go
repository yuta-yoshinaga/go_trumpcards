//go:build !js || !wasm || extra

package domain

import "encoding/json"

// FrenchTarotPlayer フレンチタロット (French Tarot) のプレイヤークラス。手札
// (GamePlayer) と獲得トリック (TrickHolder) を保持する。得点はゲーム本体が管理する。
type FrenchTarotPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewFrenchTarotPlayer コンストラクタ
func NewFrenchTarotPlayer(isHuman bool) *FrenchTarotPlayer {
	return &FrenchTarotPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット (トリック・手札・終了状態を初期化)
func (p *FrenchTarotPlayer) ResetRound() {
	resetPlayerRound(p)
}

// frenchTarotPlayerJSON is the JSON wire format for FrenchTarotPlayer.
type frenchTarotPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *FrenchTarotPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(frenchTarotPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *FrenchTarotPlayer) UnmarshalJSON(data []byte) error {
	var j frenchTarotPlayerJSON
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
