//go:build !js || !wasm || extra11

package domain

import "encoding/json"

// MarriagePlayer マリッジのプレイヤー。手札とラウンド／累計スコアを保持する。
type MarriagePlayer struct {
	*GamePlayer
	RoundScoreHolder
}

// NewMarriagePlayer コンストラクタ
func NewMarriagePlayer(isHuman bool) *MarriagePlayer {
	return &MarriagePlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・終了状態を初期化）
func (p *MarriagePlayer) ResetRound() {
	resetRoundScored(p)
}

// marriagePlayerJSON は MarriagePlayer の JSON 表現。
type marriagePlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
}

// MarshalJSON implements json.Marshaler.
func (p *MarriagePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(marriagePlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *MarriagePlayer) UnmarshalJSON(data []byte) error {
	var j marriagePlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.GamePlayer = restoreGamePlayer(j.GamePlayer)
	assignIfSet(&p.RoundScoreHolder, j.RoundScoreHolder)
	return nil
}
