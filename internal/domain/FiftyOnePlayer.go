package domain

import "encoding/json"

// FiftyOneMaxScore 最高得点 (A+K+Q+J+10 同スート = 51)
const FiftyOneMaxScore = 51

// FiftyOnePlayer フィフティワンプレイヤー
type FiftyOnePlayer struct {
	*GamePlayer
}

// NewFiftyOnePlayer コンストラクタ
func NewFiftyOnePlayer(isHuman bool) *FiftyOnePlayer {
	return &FiftyOnePlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// fiftyOneCardScore カードの51得点を返す (A=11, J/Q/K=10, 2-10=額面)
func fiftyOneCardScore(c *Card) int {
	v := c.GetValue()
	if v == 1 {
		return 11 // Ace
	}
	if v >= 11 {
		return 10 // J, Q, K
	}
	return v
}

// SuitScores スートごとの得点マップを返す
func (p *FiftyOnePlayer) SuitScores() map[int]int {
	scores := map[int]int{
		CardDesignSpade:   0,
		CardDesignClover:  0,
		CardDesignHeart:   0,
		CardDesignDiamond: 0,
	}
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		scores[c.GetDesign()] += fiftyOneCardScore(c)
	}
	return scores
}

// BestSuitScore 最高スート得点を返す
func (p *FiftyOnePlayer) BestSuitScore() int {
	best := 0
	for _, s := range p.SuitScores() {
		if s > best {
			best = s
		}
	}
	return best
}

// BestSuit 最高得点のスートを返す
func (p *FiftyOnePlayer) BestSuit() int {
	best := 0
	bestDesign := CardDesignSpade
	for design, s := range p.SuitScores() {
		if s > best {
			best = s
			bestDesign = design
		}
	}
	return bestDesign
}

// fiftyOnePlayerJSON is the JSON wire format for FiftyOnePlayer.
type fiftyOnePlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
}

// MarshalJSON implements json.Marshaler.
func (p *FiftyOnePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(fiftyOnePlayerJSON{GamePlayer: p.GamePlayer})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *FiftyOnePlayer) UnmarshalJSON(data []byte) error {
	var j fiftyOnePlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	return nil
}
