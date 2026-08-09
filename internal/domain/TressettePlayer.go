//go:build !js || !wasm || casino

package domain

import "encoding/json"

// TressettePlayer トレセッテのプレイヤークラス。手札（GamePlayer）と獲得トリック
// （TrickHolder）を保持する。得点はチーム単位で Tressette 本体が管理するため、
// プレイヤー個別の得点ホルダーは持たない。
type TressettePlayer struct {
	*GamePlayer
	TrickHolder
}

// NewTressettePlayer コンストラクタ
func NewTressettePlayer(isHuman bool) *TressettePlayer {
	return &TressettePlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *TressettePlayer) ResetRound() {
	resetPlayerRound(p)
}

// tressettePlayerJSON is the JSON wire format for TressettePlayer.
type tressettePlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *TressettePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(tressettePlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *TressettePlayer) UnmarshalJSON(data []byte) error {
	var j tressettePlayerJSON
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
