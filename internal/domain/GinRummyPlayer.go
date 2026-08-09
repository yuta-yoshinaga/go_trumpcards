package domain

import "encoding/json"

// GinRummyPlayer ジンラミープレイヤークラス
type GinRummyPlayer struct {
	*GamePlayer
	RoundScoreHolder
}

// NewGinRummyPlayer コンストラクタ
func NewGinRummyPlayer(isHuman bool) *GinRummyPlayer {
	return &GinRummyPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・終了状態を初期化）
func (p *GinRummyPlayer) ResetRound() {
	resetRoundScored(p)
}

// ginRummyPlayerJSON is the JSON wire format for GinRummyPlayer.
type ginRummyPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
}

// MarshalJSON implements json.Marshaler.
func (p *GinRummyPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(ginRummyPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *GinRummyPlayer) UnmarshalJSON(data []byte) error {
	var j ginRummyPlayerJSON
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
