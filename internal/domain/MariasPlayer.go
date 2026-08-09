//go:build !js || !wasm || classic

package domain

import "encoding/json"

// MariasPlayer マリアーシュのプレイヤークラス。手札（GamePlayer）と獲得トリック
// （TrickHolder）を保持する。点はゲーム本体が管理する。
type MariasPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewMariasPlayer コンストラクタ
func NewMariasPlayer(isHuman bool) *MariasPlayer {
	return &MariasPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *MariasPlayer) ResetRound() {
	resetPlayerRound(p)
}

// mariasPlayerJSON is the JSON wire format for MariasPlayer.
type mariasPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *MariasPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(mariasPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *MariasPlayer) UnmarshalJSON(data []byte) error {
	var j mariasPlayerJSON
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
