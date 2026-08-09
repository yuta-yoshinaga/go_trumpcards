package domain

import "encoding/json"

// CrazyEightsPlayer クレイジーエイトプレイヤークラス
type CrazyEightsPlayer struct {
	*GamePlayer
	RoundScoreHolder
}

// NewCrazyEightsPlayer コンストラクタ
func NewCrazyEightsPlayer(isHuman bool) *CrazyEightsPlayer {
	return &CrazyEightsPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・終了状態を初期化）
func (p *CrazyEightsPlayer) ResetRound() {
	resetRoundScored(p)
}

// crazyEightsPlayerJSON is the JSON wire format for CrazyEightsPlayer.
type crazyEightsPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
}

// MarshalJSON implements json.Marshaler.
func (p *CrazyEightsPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(crazyEightsPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *CrazyEightsPlayer) UnmarshalJSON(data []byte) error {
	var j crazyEightsPlayerJSON
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
