//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// PreferencePlayer プレフェランスのプレイヤークラス。手札（GamePlayer）と獲得
// トリック（TrickHolder）を保持する。点はゲーム本体が管理する。
type PreferencePlayer struct {
	*GamePlayer
	TrickHolder
}

// NewPreferencePlayer コンストラクタ
func NewPreferencePlayer(isHuman bool) *PreferencePlayer {
	return &PreferencePlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *PreferencePlayer) ResetRound() {
	resetPlayerRound(p)
}

// preferencePlayerJSON is the JSON wire format for PreferencePlayer.
type preferencePlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *PreferencePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(preferencePlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *PreferencePlayer) UnmarshalJSON(data []byte) error {
	var j preferencePlayerJSON
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
