//go:build !js || !wasm || solo

package domain

import "encoding/json"

// TarocchiniPlayer タロッキーニのプレイヤークラス。手札 (GamePlayer) と獲得トリック
// (TrickHolder) を保持する。得点はチーム単位でゲーム本体が管理する。
type TarocchiniPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewTarocchiniPlayer コンストラクタ
func NewTarocchiniPlayer(isHuman bool) *TarocchiniPlayer {
	return &TarocchiniPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット (トリック・手札・終了状態を初期化)
func (p *TarocchiniPlayer) ResetRound() {
	resetPlayerRound(p)
}

// tarocchiniPlayerJSON is the JSON wire format for TarocchiniPlayer.
type tarocchiniPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *TarocchiniPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(tarocchiniPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *TarocchiniPlayer) UnmarshalJSON(data []byte) error {
	var j tarocchiniPlayerJSON
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
