//go:build !js || !wasm || extra

package domain

import "encoding/json"

// CostlyColoursPlayer はコストリー・カラーズのプレイヤー。
type CostlyColoursPlayer struct {
	*GamePlayer
	// score は通算得点。**ペグボードは使わないので単なる整数。**
	score int
	// played はこのディールで出した札。ショーではこれと手札を合わせて数える。
	played []*Card
	// moggedIn は交換で受け取った札があるか。
	moggedIn bool
}

// NewCostlyColoursPlayer constructs a CostlyColoursPlayer.
func NewCostlyColoursPlayer(isHuman bool) *CostlyColoursPlayer {
	return &CostlyColoursPlayer{GamePlayer: NewGamePlayer(isHuman), played: make([]*Card, 0, 4)}
}

// GetHand は手札を並べて返す。
func (p *CostlyColoursPlayer) GetHand() []*Card {
	n := p.GetCardsSize()
	out := make([]*Card, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, p.GetCard(i))
	}
	return out
}

// GetPlayed はこのディールで出した札を返す。
func (p *CostlyColoursPlayer) GetPlayed() []*Card { return p.played }

// AddPlayed は出した札を記録する。
//
// **ショーは出し切ったあとに数える。** 手札は空になるので、出した札を
// 覚えておかないと色とスートの役が判定できない。
func (p *CostlyColoursPlayer) AddPlayed(c *Card) {
	if c != nil {
		p.played = append(p.played, c)
	}
}

// IsMoggedIn は交換で札を受け取ったかを返す。
func (p *CostlyColoursPlayer) IsMoggedIn() bool { return p.moggedIn }

// SetMoggedIn は交換で札を受け取ったかを設定する。
func (p *CostlyColoursPlayer) SetMoggedIn(v bool) { p.moggedIn = v }

// GetScore は通算得点を返す。
func (p *CostlyColoursPlayer) GetScore() int { return p.score }

// AddScore は得点を加える。
func (p *CostlyColoursPlayer) AddScore(n int) { p.score += n }

// ResetScore は通算得点を 0 に戻す。
func (p *CostlyColoursPlayer) ResetScore() { p.score = 0 }

// ResetDeal は 1 ディール分の状態をクリアする。**通算得点は残す。**
func (p *CostlyColoursPlayer) ResetDeal() {
	p.Reset()
	p.played = make([]*Card, 0, 4)
	p.moggedIn = false
	p.SetIsFinished(false)
}

// costlyColoursPlayerJSON is the JSON wire format for CostlyColoursPlayer.
//
// **非公開フィールドだけの型は MarshalJSON が無いと `{}` になる。** 出した札が
// 消えると、復元した盤ではショーが数えられなくなる。
type costlyColoursPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Score      int         `json:"sc"`
	Played     []*Card     `json:"pd"`
	MoggedIn   bool        `json:"mi"`
}

// MarshalJSON implements json.Marshaler.
func (p *CostlyColoursPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(costlyColoursPlayerJSON{
		GamePlayer: p.GamePlayer, Score: p.score, Played: p.played, MoggedIn: p.moggedIn,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *CostlyColoursPlayer) UnmarshalJSON(data []byte) error {
	var j costlyColoursPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.score = j.Score
	p.played = j.Played
	if p.played == nil {
		p.played = make([]*Card, 0, 4)
	}
	p.moggedIn = j.MoggedIn
	return nil
}
