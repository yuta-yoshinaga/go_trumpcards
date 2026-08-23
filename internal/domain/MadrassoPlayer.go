//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// MadrassoPlayer マドラッソのプレイヤークラス。手札（GamePlayer）と獲得トリック
// （TrickHolder）を保持する。得点はチーム単位で Madrasso 本体が管理するため、
// プレイヤー個別の得点ホルダーは持たない。
type MadrassoPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewMadrassoPlayer コンストラクタ
func NewMadrassoPlayer(isHuman bool) *MadrassoPlayer {
	return &MadrassoPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *MadrassoPlayer) ResetRound() {
	resetPlayerRound(p)
}

// madrassoPlayerJSON is the JSON wire format for MadrassoPlayer.
type madrassoPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *MadrassoPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(madrassoPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *MadrassoPlayer) UnmarshalJSON(data []byte) error {
	var j madrassoPlayerJSON
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
