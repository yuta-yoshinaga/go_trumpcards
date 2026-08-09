//go:build !js || !wasm || classic

package domain

import "encoding/json"

// SoloWhistPlayer ソロ・ホイストのプレイヤークラス。手札（GamePlayer）と獲得
// トリック（TrickHolder）を保持する。点はゲーム本体が管理する。
type SoloWhistPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewSoloWhistPlayer コンストラクタ
func NewSoloWhistPlayer(isHuman bool) *SoloWhistPlayer {
	return &SoloWhistPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *SoloWhistPlayer) ResetRound() {
	resetPlayerRound(p)
}

// soloWhistPlayerJSON is the JSON wire format for SoloWhistPlayer.
type soloWhistPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *SoloWhistPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(soloWhistPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SoloWhistPlayer) UnmarshalJSON(data []byte) error {
	var j soloWhistPlayerJSON
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
