//go:build !js || !wasm || extra

package domain

import "encoding/json"

// ViraPlayer ヴィーラのプレイヤークラス。手札（GamePlayer）と獲得
// トリック（TrickHolder）を保持する。点はゲーム本体が管理する。
type ViraPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewViraPlayer コンストラクタ
func NewViraPlayer(isHuman bool) *ViraPlayer {
	return &ViraPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *ViraPlayer) ResetRound() {
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
}

// viraPlayerJSON is the JSON wire format for ViraPlayer.
type viraPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *ViraPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(viraPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ViraPlayer) UnmarshalJSON(data []byte) error {
	var j viraPlayerJSON
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
