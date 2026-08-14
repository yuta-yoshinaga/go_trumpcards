//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// GoofspielPlayer はゴフスピールのプレイヤー。
//
// **手札は最初から最後まで自分のスート 13 枚**で、使った札だけが減ります。
// 相手の手札の中身は「まだ出していない札」として完全に分かる——隠されているのは
// **今このラウンドで何を出したか**だけです。
type GoofspielPlayer struct {
	*GamePlayer
	score int
}

// NewGoofspielPlayer はコンストラクタ。
func NewGoofspielPlayer(isHuman bool) *GoofspielPlayer {
	return &GoofspielPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// GetScore は得点を返す。
func (p *GoofspielPlayer) GetScore() int { return p.score }

// SetScore は得点を設定する (主にテスト/復元用)。
func (p *GoofspielPlayer) SetScore(n int) { p.score = n }

// AddScore は得点を加える。
func (p *GoofspielPlayer) AddScore(n int) { p.score += n }

// ResetGame はゲーム開始時の状態に戻す。
func (p *GoofspielPlayer) ResetGame() {
	p.Reset()
	p.score = 0
	p.SetIsFinished(false)
}

// goofspielPlayerJSON is the JSON wire format for GoofspielPlayer.
type goofspielPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Score      int         `json:"sc"`
}

// MarshalJSON implements json.Marshaler.
func (p *GoofspielPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(goofspielPlayerJSON{GamePlayer: p.GamePlayer, Score: p.score})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *GoofspielPlayer) UnmarshalJSON(data []byte) error {
	var j goofspielPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer == nil {
		return errors.New("goofspiel player is missing its base player")
	}
	// **得点は賞札のランクの合計。** 13 枚ぶんを超えることも、負になることもありません。
	const maxScore = GoofspielRounds * (GoofspielRounds + 1) / 2
	if j.Score < 0 || j.Score > maxScore {
		return fmt.Errorf("score must be between 0 and %d, got %d", maxScore, j.Score)
	}
	p.GamePlayer = j.GamePlayer
	p.score = j.Score
	return nil
}
