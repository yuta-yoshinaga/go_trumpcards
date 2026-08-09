//go:build !js || !wasm || classic

package domain

import "encoding/json"

// ManillePlayer マニーユのプレイヤークラス。手札（GamePlayer）と獲得トリック
// （TrickHolder）を保持する。点はチーム単位で Manille 本体が管理する。
type ManillePlayer struct {
	*GamePlayer
	TrickHolder
}

// NewManillePlayer コンストラクタ
func NewManillePlayer(isHuman bool) *ManillePlayer {
	return &ManillePlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *ManillePlayer) ResetRound() {
	resetPlayerRound(p)
}

// manillePlayerJSON is the JSON wire format for ManillePlayer.
type manillePlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *ManillePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(manillePlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ManillePlayer) UnmarshalJSON(data []byte) error {
	var j manillePlayerJSON
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
