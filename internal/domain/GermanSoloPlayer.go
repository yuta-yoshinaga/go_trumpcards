//go:build !js || !wasm || classic

package domain

import "encoding/json"

// GermanSoloPlayer ジャーマン・ソロ (GermanSolo) のプレイヤークラス。手札 (GamePlayer) と
// 獲得トリック (TrickHolder) を保持する。点はゲーム本体が管理する。
type GermanSoloPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewGermanSoloPlayer コンストラクタ
func NewGermanSoloPlayer(isHuman bool) *GermanSoloPlayer {
	return &GermanSoloPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット (トリック・手札・終了状態を初期化)
func (p *GermanSoloPlayer) ResetRound() {
	resetPlayerRound(p)
}

// germanSoloPlayerJSON is the JSON wire format for GermanSoloPlayer.
type germanSoloPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *GermanSoloPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(germanSoloPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *GermanSoloPlayer) UnmarshalJSON(data []byte) error {
	var j germanSoloPlayerJSON
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
