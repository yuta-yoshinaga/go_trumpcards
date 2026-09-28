//go:build !js || !wasm || extra7

package domain

import "encoding/json"

// HeartsPlayer ハーツプレイヤークラス
type HeartsPlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
}

// NewHeartsPlayer コンストラクタ
func NewHeartsPlayer(isHuman bool) *HeartsPlayer {
	return &HeartsPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（スコア・トリック・手札・終了状態を初期化）
func (p *HeartsPlayer) ResetRound() {
	resetRoundWithTricks(p)
}

// heartsPlayerJSON is the JSON wire format for HeartsPlayer.
type heartsPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *HeartsPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(heartsPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *HeartsPlayer) UnmarshalJSON(data []byte) error {
	var j heartsPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.GamePlayer = restoreGamePlayer(j.GamePlayer)
	assignIfSet(&p.RoundScoreHolder, j.RoundScoreHolder)
	assignIfSet(&p.TrickHolder, j.TrickHolder)
	return nil
}
