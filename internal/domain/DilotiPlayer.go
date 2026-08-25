//go:build !js || !wasm || classic

package domain

import "encoding/json"

// DilotiPlayer はディロティのプレイヤー。
type DilotiPlayer struct {
	*GamePlayer
	// captured はこの局で取った札。
	captured []*Card
	// xeri はこの局のクセリ回数。**1 回 10 点**なので、枚数より重い。
	xeri int
	// score は通算得点。
	score int
}

// NewDilotiPlayer constructs a DilotiPlayer.
func NewDilotiPlayer(isHuman bool) *DilotiPlayer {
	return &DilotiPlayer{GamePlayer: NewGamePlayer(isHuman), captured: make([]*Card, 0, DilotiDeckSize)}
}

// GetHand は手札を並べて返す。
//
// **Player は枚数と 1 枚ずつしか公開していない。** 捕獲の列挙は手札全体を
// 一度に見るので、ここで並べ直す。
func (p *DilotiPlayer) GetHand() []*Card {
	n := p.GetCardsSize()
	out := make([]*Card, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, p.GetCard(i))
	}
	return out
}

// GetCaptured はこの局で取った札を返す。
func (p *DilotiPlayer) GetCaptured() []*Card { return p.captured }

// AddCaptured は取った札を積む。
func (p *DilotiPlayer) AddCaptured(cards []*Card) {
	for _, c := range cards {
		if c != nil {
			p.captured = append(p.captured, c)
		}
	}
}

// GetXeri はこの局のクセリ回数を返す。
func (p *DilotiPlayer) GetXeri() int { return p.xeri }

// AddXeri はクセリを 1 回加える。
func (p *DilotiPlayer) AddXeri() { p.xeri++ }

// GetScore は通算得点を返す。
func (p *DilotiPlayer) GetScore() int { return p.score }

// AddScore は得点を加える。
func (p *DilotiPlayer) AddScore(n int) { p.score += n }

// ResetScore は通算得点を 0 に戻す。
func (p *DilotiPlayer) ResetScore() { p.score = 0 }

// ResetRound は 1 局分の状態をクリアする。**通算得点は残す。**
func (p *DilotiPlayer) ResetRound() {
	p.Reset()
	p.captured = make([]*Card, 0, DilotiDeckSize)
	p.xeri = 0
	p.SetIsFinished(false)
}

// dilotiPlayerJSON is the JSON wire format for DilotiPlayer.
//
// **非公開フィールドだけの型は MarshalJSON が無いと `{}` になる。** 取り札と
// クセリ回数が消えると、復元した盤で集計だけが 0 になって静かに壊れる。
type dilotiPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Captured   []*Card     `json:"cp"`
	Xeri       int         `json:"xr"`
	Score      int         `json:"sc"`
}

// MarshalJSON implements json.Marshaler.
func (p *DilotiPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(dilotiPlayerJSON{
		GamePlayer: p.GamePlayer, Captured: p.captured, Xeri: p.xeri, Score: p.score,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *DilotiPlayer) UnmarshalJSON(data []byte) error {
	var j dilotiPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.captured = j.Captured
	if p.captured == nil {
		p.captured = make([]*Card, 0, DilotiDeckSize)
	}
	p.xeri = j.Xeri
	p.score = j.Score
	return nil
}
