//go:build !js || !wasm || casino

package domain

import "encoding/json"

// EcartePlayerCnt Écarté プレイヤー数 (2人固定)
const EcartePlayerCnt = 2

// EcartePlayer Écarté プレイヤークラス
type EcartePlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
}

// NewEcartePlayer コンストラクタ
func NewEcartePlayer(isHuman bool) *EcartePlayer {
	return &EcartePlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetGame ゲーム単位の状態をリセット
func (p *EcartePlayer) ResetGame() {
	p.SetRoundScore(0)
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
}

// ecartePlayerJSON is the JSON wire format for EcartePlayer.
type ecartePlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *EcartePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(ecartePlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *EcartePlayer) UnmarshalJSON(data []byte) error {
	var j ecartePlayerJSON
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
