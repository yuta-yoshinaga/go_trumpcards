//go:build !js || !wasm || extra6

package domain

import "encoding/json"

// CribbagePlayer クリベッジプレイヤークラス
type CribbagePlayer struct {
	*GamePlayer
	RoundScoreHolder
}

// cribbagePlayerJSON is the JSON wire format for CribbagePlayer.
type cribbagePlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
}

// MarshalJSON implements json.Marshaler.
func (p *CribbagePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(cribbagePlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *CribbagePlayer) UnmarshalJSON(data []byte) error {
	var j cribbagePlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.GamePlayer = restoreGamePlayer(j.GamePlayer)
	assignIfSet(&p.RoundScoreHolder, j.RoundScoreHolder)
	return nil
}

// NewCribbagePlayer コンストラクタ
func NewCribbagePlayer(isHuman bool) *CribbagePlayer {
	return &CribbagePlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・終了状態を初期化）
func (p *CribbagePlayer) ResetRound() {
	resetRoundScored(p)
}
