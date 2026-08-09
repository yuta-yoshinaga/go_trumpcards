//go:build !js || !wasm || casino

package domain

import "encoding/json"

// TutePlayer トゥーテのプレイヤークラス。手札（GamePlayer）と獲得トリック
// （TrickHolder）を保持する。点はチーム単位で Tute 本体が管理する。
type TutePlayer struct {
	*GamePlayer
	TrickHolder
}

// NewTutePlayer コンストラクタ
func NewTutePlayer(isHuman bool) *TutePlayer {
	return &TutePlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *TutePlayer) ResetRound() {
	resetPlayerRound(p)
}

// tutePlayerJSON is the JSON wire format for TutePlayer.
type tutePlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *TutePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(tutePlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *TutePlayer) UnmarshalJSON(data []byte) error {
	var j tutePlayerJSON
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
