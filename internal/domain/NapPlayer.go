//go:build !js || !wasm || classic

package domain

import "encoding/json"

// NapPlayer ナップのプレイヤークラス。手札（GamePlayer）と獲得トリック
// （TrickHolder）を保持する。チップはゲーム本体が管理する。
type NapPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewNapPlayer コンストラクタ
func NewNapPlayer(isHuman bool) *NapPlayer {
	return &NapPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *NapPlayer) ResetRound() {
	resetPlayerRound(p)
}

// napPlayerJSON is the JSON wire format for NapPlayer.
type napPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *NapPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(napPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *NapPlayer) UnmarshalJSON(data []byte) error {
	var j napPlayerJSON
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
