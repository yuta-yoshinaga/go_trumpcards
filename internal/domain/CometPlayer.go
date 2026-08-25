//go:build !js || !wasm || solo

package domain

import "encoding/json"

// CometPlayer はコメットのプレイヤー。
type CometPlayer struct {
	*GamePlayer
	// score は通算得点。**勝ち抜けではなく先に目標点へ届いた側の勝ち。**
	score int
}

// NewCometPlayer constructs a CometPlayer.
func NewCometPlayer(isHuman bool) *CometPlayer {
	return &CometPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// GetHand は手札を並べて返す。
//
// **Player は枚数と 1 枚ずつしか公開していない。** 出せる札の列挙は手札全体を
// 一度に見るので、ここで並べ直す。
func (p *CometPlayer) GetHand() []*Card {
	n := p.GetCardsSize()
	out := make([]*Card, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, p.GetCard(i))
	}
	return out
}

// GetScore は通算得点を返す。
func (p *CometPlayer) GetScore() int { return p.score }

// AddScore は得点を加える。
func (p *CometPlayer) AddScore(n int) { p.score += n }

// ResetScore は通算得点を 0 に戻す。
func (p *CometPlayer) ResetScore() { p.score = 0 }

// ResetRound は 1 局分の状態をクリアする。**通算得点は残す。**
func (p *CometPlayer) ResetRound() {
	p.Reset()
	p.SetIsFinished(false)
}

// cometPlayerJSON is the JSON wire format for CometPlayer.
//
// **非公開フィールドだけの型は MarshalJSON が無いと `{}` になる。** 得点が
// 消えると、復元した盤で勝敗だけが決まらなくなる。
type cometPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Score      int         `json:"sc"`
}

// MarshalJSON implements json.Marshaler.
func (p *CometPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(cometPlayerJSON{GamePlayer: p.GamePlayer, Score: p.score})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *CometPlayer) UnmarshalJSON(data []byte) error {
	var j cometPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.score = j.Score
	return nil
}
