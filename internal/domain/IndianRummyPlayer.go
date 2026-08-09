//go:build !js || !wasm || extra

package domain

import "encoding/json"

// IndianRummyPlayer インドラミーのプレイヤー。手札とラウンド／累計スコアを保持する。
type IndianRummyPlayer struct {
	*GamePlayer
	RoundScoreHolder
}

// NewIndianRummyPlayer コンストラクタ
func NewIndianRummyPlayer(isHuman bool) *IndianRummyPlayer {
	return &IndianRummyPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・終了状態を初期化）
func (p *IndianRummyPlayer) ResetRound() {
	resetRoundScored(p)
}

// indianRummyPlayerJSON は IndianRummyPlayer の JSON 表現。
type indianRummyPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
}

// MarshalJSON implements json.Marshaler.
func (p *IndianRummyPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(indianRummyPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *IndianRummyPlayer) UnmarshalJSON(data []byte) error {
	var j indianRummyPlayerJSON
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
