//go:build !js || !wasm || solo

package domain

import "encoding/json"

// YanivCallThreshold Yaniv 宣言が可能な手札合計の上限
const YanivCallThreshold = 5

// YanivAsafPenalty アサフ成立時に宣言者へ加算されるペナルティ点
const YanivAsafPenalty = 30

// YanivPlayer Yaniv プレイヤー
type YanivPlayer struct {
	*GamePlayer
	score      int  // 累計失点 (低いほど良い)
	eliminated bool // 脱落済みか
}

// NewYanivPlayer コンストラクタ
func NewYanivPlayer(isHuman bool) *YanivPlayer {
	return &YanivPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// yanivCardValue カードの失点を返す (ジョーカー=0, A=1, J/Q/K=10, 2-10=額面)
func yanivCardValue(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetDesign() == CardDesignJoker {
		return 0
	}
	v := c.GetValue()
	if v >= 10 { // 10, J, Q, K
		return 10
	}
	return v // A=1, 2-9 は額面
}

// HandTotal 手札の失点合計を返す
func (p *YanivPlayer) HandTotal() int {
	total := 0
	for i := range p.GetCardsSize() {
		total += yanivCardValue(p.GetCard(i))
	}
	return total
}

// GetScore 累計失点を返す
func (p *YanivPlayer) GetScore() int { return p.score }

// SetScore 累計失点を設定する
func (p *YanivPlayer) SetScore(n int) { p.score = n }

// AddScore 累計失点に加算する
func (p *YanivPlayer) AddScore(n int) { p.score += n }

// IsEliminated 脱落済みかを返す
func (p *YanivPlayer) IsEliminated() bool { return p.eliminated }

// SetEliminated 脱落状態を設定する
func (p *YanivPlayer) SetEliminated(v bool) { p.eliminated = v }

// yanivPlayerJSON is the JSON wire format for YanivPlayer.
type yanivPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Score      int         `json:"sc"`
	Eliminated bool        `json:"el"`
}

// MarshalJSON implements json.Marshaler.
func (p *YanivPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(yanivPlayerJSON{GamePlayer: p.GamePlayer, Score: p.score, Eliminated: p.eliminated})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *YanivPlayer) UnmarshalJSON(data []byte) error {
	var j yanivPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.score = j.Score
	p.eliminated = j.Eliminated
	return nil
}
