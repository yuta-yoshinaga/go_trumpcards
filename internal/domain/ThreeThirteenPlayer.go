//go:build !js || !wasm || extra

package domain

import "encoding/json"

// ThreeThirteenPlayer は Three Thirteen のプレイヤー。手札・累計／ラウンドスコアを保持する。
type ThreeThirteenPlayer struct {
	*GamePlayer
	RoundScoreHolder
}

// NewThreeThirteenPlayer コンストラクタ
func NewThreeThirteenPlayer(isHuman bool) *ThreeThirteenPlayer {
	return &ThreeThirteenPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（手札・ラウンドスコア・終了状態を初期化）
func (p *ThreeThirteenPlayer) ResetRound() {
	resetRoundScored(p)
}

// threeThirteenPlayerJSON は ThreeThirteenPlayer の JSON 表現。
type threeThirteenPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
}

// MarshalJSON implements json.Marshaler.
func (p *ThreeThirteenPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(threeThirteenPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ThreeThirteenPlayer) UnmarshalJSON(data []byte) error {
	var j threeThirteenPlayerJSON
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
