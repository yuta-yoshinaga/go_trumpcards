//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// TrappolaPlayer トラッポラのプレイヤークラス。手札（GamePlayer）と獲得トリック
// （TrickHolder）を保持する。得点はチーム単位で Trappola 本体が管理するため、
// プレイヤー個別の得点ホルダーは持たない。
type TrappolaPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewTrappolaPlayer コンストラクタ
func NewTrappolaPlayer(isHuman bool) *TrappolaPlayer {
	return &TrappolaPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *TrappolaPlayer) ResetRound() {
	resetPlayerRound(p)
}

// trappolaPlayerJSON is the JSON wire format for TrappolaPlayer.
type trappolaPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *TrappolaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(trappolaPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *TrappolaPlayer) UnmarshalJSON(data []byte) error {
	var j trappolaPlayerJSON
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
