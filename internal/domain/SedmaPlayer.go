//go:build !js || !wasm || classic

package domain

import "encoding/json"

// SedmaPlayer セドマのプレイヤークラス。手札（GamePlayer）と獲得トリック
// （TrickHolder）を保持する。点はチーム単位で Sedma 本体が管理する。
type SedmaPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewSedmaPlayer コンストラクタ
func NewSedmaPlayer(isHuman bool) *SedmaPlayer {
	return &SedmaPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *SedmaPlayer) ResetRound() {
	resetPlayerRound(p)
}

// sedmaPlayerJSON is the JSON wire format for SedmaPlayer.
type sedmaPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *SedmaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(sedmaPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SedmaPlayer) UnmarshalJSON(data []byte) error {
	var j sedmaPlayerJSON
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
