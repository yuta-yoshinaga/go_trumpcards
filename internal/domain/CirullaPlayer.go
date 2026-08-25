//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// CirullaPlayer はチルッラのプレイヤー。
type CirullaPlayer struct {
	*GamePlayer
	// captured はこのラウンドで取った札。
	captured []*Card
	// scope はこのラウンドのスコパ回数。
	scope int
	// bonusPoints は配札ボーナスで得た点。
	bonusPoints int
	// score はマッチ通算の点。
	score int
}

// NewCirullaPlayer constructs a CirullaPlayer.
func NewCirullaPlayer(isHuman bool) *CirullaPlayer {
	return &CirullaPlayer{GamePlayer: NewGamePlayer(isHuman), captured: make([]*Card, 0)}
}

// GetCaptured は取った札を返す。
func (p *CirullaPlayer) GetCaptured() []*Card { return p.captured }

// AddCaptured は取った札を加える。
func (p *CirullaPlayer) AddCaptured(cards []*Card) {
	for _, c := range cards {
		if c != nil {
			p.captured = append(p.captured, c)
		}
	}
}

// GetScope はスコパ回数を返す。
func (p *CirullaPlayer) GetScope() int { return p.scope }

// AddScopa はスコパを 1 回加える。
func (p *CirullaPlayer) AddScopa() { p.scope++ }

// GetBonusPoints は配札ボーナスの点を返す。
func (p *CirullaPlayer) GetBonusPoints() int { return p.bonusPoints }

// AddBonusPoints は配札ボーナスの点を加える。
func (p *CirullaPlayer) AddBonusPoints(n int) { p.bonusPoints += n }

// GetScore はマッチ通算の点を返す。
func (p *CirullaPlayer) GetScore() int { return p.score }

// AddScore はマッチ通算の点を加える。
func (p *CirullaPlayer) AddScore(n int) { p.score += n }

// ResetScore はマッチ通算の点を 0 に戻す。
func (p *CirullaPlayer) ResetScore() { p.score = 0 }

// ResetRound は 1 ラウンド分の状態をクリアする。通算点は維持する。
func (p *CirullaPlayer) ResetRound() {
	p.Reset()
	p.captured = make([]*Card, 0)
	p.scope = 0
	p.bonusPoints = 0
	p.SetIsFinished(false)
}

// cirullaPlayerJSON is the JSON wire format for CirullaPlayer.
type cirullaPlayerJSON struct {
	GamePlayer  *GamePlayer `json:"gp"`
	Captured    []*Card     `json:"cp"`
	Scope       int         `json:"sc"`
	BonusPoints int         `json:"bp"`
	Score       int         `json:"sr"`
}

// MarshalJSON implements json.Marshaler.
func (p *CirullaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(cirullaPlayerJSON{
		GamePlayer:  p.GamePlayer,
		Captured:    p.captured,
		Scope:       p.scope,
		BonusPoints: p.bonusPoints,
		Score:       p.score,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *CirullaPlayer) UnmarshalJSON(data []byte) error {
	var j cirullaPlayerJSON
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
		p.captured = make([]*Card, 0)
	}
	p.scope = j.Scope
	p.bonusPoints = j.BonusPoints
	p.score = j.Score
	return nil
}
