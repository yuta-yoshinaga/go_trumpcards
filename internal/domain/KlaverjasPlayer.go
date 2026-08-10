//go:build !js || !wasm || classic

package domain

import "encoding/json"

// KlaverjasPlayer クラヴァヤスのプレイヤークラス。手札（GamePlayer）と獲得
// トリック（TrickHolder）を保持する。点はチーム単位で Klaverjas 本体が管理する。
type KlaverjasPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewKlaverjasPlayer コンストラクタ
func NewKlaverjasPlayer(isHuman bool) *KlaverjasPlayer {
	return &KlaverjasPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *KlaverjasPlayer) ResetRound() {
	resetPlayerRound(p)
}

// klaverjasPlayerJSON is the JSON wire format for KlaverjasPlayer.
type klaverjasPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *KlaverjasPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(klaverjasPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *KlaverjasPlayer) UnmarshalJSON(data []byte) error {
	var j klaverjasPlayerJSON
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
