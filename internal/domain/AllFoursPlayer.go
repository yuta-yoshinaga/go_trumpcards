package domain

import "encoding/json"

// AllFoursPlayer All Fours (Seven Up) プレイヤークラス
type AllFoursPlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
}

// NewAllFoursPlayer コンストラクタ
func NewAllFoursPlayer(isHuman bool) *AllFoursPlayer {
	return &AllFoursPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット (得点・トリック・手札・終了状態を初期化)
func (p *AllFoursPlayer) ResetRound() {
	resetRoundWithTricks(p)
}

// allFoursPlayerJSON is the JSON wire format for AllFoursPlayer.
type allFoursPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *AllFoursPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(allFoursPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *AllFoursPlayer) UnmarshalJSON(data []byte) error {
	var j allFoursPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.GamePlayer = restoreGamePlayer(j.GamePlayer)
	assignIfSet(&p.RoundScoreHolder, j.RoundScoreHolder)
	assignIfSet(&p.TrickHolder, j.TrickHolder)
	return nil
}
