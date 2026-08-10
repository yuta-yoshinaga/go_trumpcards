//go:build !js || !wasm || extra

package domain

import "encoding/json"

// MachiavelliPlayer マキャヴェッリのプレイヤー。手札とラウンド／累計スコアを保持する。
type MachiavelliPlayer struct {
	*GamePlayer
	RoundScoreHolder
}

// NewMachiavelliPlayer コンストラクタ
func NewMachiavelliPlayer(isHuman bool) *MachiavelliPlayer {
	return &MachiavelliPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・終了状態を初期化）
func (p *MachiavelliPlayer) ResetRound() {
	resetRoundScored(p)
}

// machiavelliPlayerJSON は MachiavelliPlayer の JSON 表現。
type machiavelliPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
}

// MarshalJSON implements json.Marshaler.
func (p *MachiavelliPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(machiavelliPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *MachiavelliPlayer) UnmarshalJSON(data []byte) error {
	var j machiavelliPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.RoundScoreHolder != nil {
		p.RoundScoreHolder = *j.RoundScoreHolder
	}
	return nil
}
