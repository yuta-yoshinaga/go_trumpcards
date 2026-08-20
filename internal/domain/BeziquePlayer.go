//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// BeziquePlayerCnt Bezique プレイヤー数 (2人固定)
const BeziquePlayerCnt = 2

// BeziquePlayer Bezique プレイヤークラス
type BeziquePlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
}

// NewBeziquePlayer コンストラクタ
func NewBeziquePlayer(isHuman bool) *BeziquePlayer {
	return &BeziquePlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetGame ゲーム単位の状態をリセット
func (p *BeziquePlayer) ResetGame() {
	p.SetRoundScore(0)
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
}

// beziquePlayerJSON is the JSON wire format for BeziquePlayer.
type beziquePlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *BeziquePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(beziquePlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BeziquePlayer) UnmarshalJSON(data []byte) error {
	var j beziquePlayerJSON
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
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	return nil
}
