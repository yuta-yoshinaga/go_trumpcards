//go:build !js || !wasm || casino

package domain

import "encoding/json"

// SuecaPlayer スエカのプレイヤークラス。手札（GamePlayer）と獲得トリック
// （TrickHolder）を保持する。点はチーム単位で Sueca 本体が管理する。
type SuecaPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewSuecaPlayer コンストラクタ
func NewSuecaPlayer(isHuman bool) *SuecaPlayer {
	return &SuecaPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *SuecaPlayer) ResetRound() {
	resetPlayerRound(p)
}

// suecaPlayerJSON is the JSON wire format for SuecaPlayer.
type suecaPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *SuecaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(suecaPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SuecaPlayer) UnmarshalJSON(data []byte) error {
	var j suecaPlayerJSON
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
