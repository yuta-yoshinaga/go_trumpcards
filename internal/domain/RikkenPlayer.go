//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
)

// RikkenPlayer はリッケンのプレイヤー。
//
// 手札 (GamePlayer) と獲得したトリック (TrickHolder)、それに通算得点を持ちます。
// **組は契約ごとに変わる**ので、チームは席ではなく本体側が持ちます。
type RikkenPlayer struct {
	*GamePlayer
	TrickHolder
	score int
}

// NewRikkenPlayer はコンストラクタ。
func NewRikkenPlayer(isHuman bool) *RikkenPlayer {
	return &RikkenPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// GetScore は通算得点を返す。
func (p *RikkenPlayer) GetScore() int { return p.score }

// AddScore は得点を加える。**負にもなります。**
func (p *RikkenPlayer) AddScore(n int) { p.score += n }

// SetScore は得点を設定する (主にテスト/復元用)。
func (p *RikkenPlayer) SetScore(n int) { p.score = n }

// ResetRound はラウンドの状態に戻す (得点は保持)。
func (p *RikkenPlayer) ResetRound() {
	resetPlayerRound(p)
}

// ResetGame はゲーム開始時の状態に戻す。
func (p *RikkenPlayer) ResetGame() {
	p.ResetRound()
	p.score = 0
}

// rikkenPlayerJSON is the JSON wire format for RikkenPlayer.
type rikkenPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Score       int          `json:"sc"`
}

// MarshalJSON implements json.Marshaler.
func (p *RikkenPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(rikkenPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Score:       p.score,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *RikkenPlayer) UnmarshalJSON(data []byte) error {
	var j rikkenPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer == nil {
		return errors.New("rikken player is missing its base player")
	}
	p.GamePlayer = j.GamePlayer
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	p.score = j.Score
	return nil
}
