package domain

import "encoding/json"

// TonkPlayer Tonkプレイヤークラス
type TonkPlayer struct {
	*GamePlayer
	RoundScoreHolder
}

// NewTonkPlayer コンストラクタ
func NewTonkPlayer(isHuman bool) *TonkPlayer {
	return &TonkPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・終了状態を初期化）
func (p *TonkPlayer) ResetRound() {
	resetRoundScored(p)
}

// tonkPlayerJSON is the JSON wire format for TonkPlayer.
type tonkPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
}

// MarshalJSON implements json.Marshaler.
func (p *TonkPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(tonkPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *TonkPlayer) UnmarshalJSON(data []byte) error {
	var j tonkPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.GamePlayer = restoreGamePlayer(j.GamePlayer)
	assignIfSet(&p.RoundScoreHolder, j.RoundScoreHolder)
	return nil
}
