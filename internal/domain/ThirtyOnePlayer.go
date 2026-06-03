//go:build !js || !wasm || solo

package domain

import "encoding/json"

// ThirtyOneTarget 勝利となるスート合計 (即ラウンド勝利)
const ThirtyOneTarget = 31

// ThirtyOnePlayer ThirtyOne (サーティワン / Scat) プレイヤー
type ThirtyOnePlayer struct {
	*GamePlayer
	lives int
}

// NewThirtyOnePlayer コンストラクタ
func NewThirtyOnePlayer(isHuman bool) *ThirtyOnePlayer {
	return &ThirtyOnePlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// thirtyOneCardScore カードの得点を返す (A=11, J/Q/K=10, 2-10=額面)
func thirtyOneCardScore(c *Card) int {
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
func (p *ThirtyOnePlayer) SuitScores() map[int]int {
	scores := map[int]int{
		CardDesignSpade:   0,
		CardDesignClover:  0,
		CardDesignHeart:   0,
		CardDesignDiamond: 0,
	}
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		scores[c.GetDesign()] += thirtyOneCardScore(c)
	}
	return scores
}

// BestSuitScore 最高スート得点を返す
func (p *ThirtyOnePlayer) BestSuitScore() int {
	best := 0
	for _, s := range p.SuitScores() {
		if s > best {
			best = s
		}
	}
	return best
}

// BestSuit 最高得点のスートを返す (決定論的順序で走査)
func (p *ThirtyOnePlayer) BestSuit() int {
	best := 0
	bestDesign := CardDesignSpade
	scores := p.SuitScores()
	for _, design := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if s := scores[design]; s > best {
			best = s
			bestDesign = design
		}
	}
	return bestDesign
}

// GetLives 残りライフ数を返す
func (p *ThirtyOnePlayer) GetLives() int { return p.lives }

// SetLives ライフ数を設定する
func (p *ThirtyOnePlayer) SetLives(n int) { p.lives = n }

// LoseLife ライフを 1 減らす
func (p *ThirtyOnePlayer) LoseLife() { p.lives-- }

// IsEliminated 脱落済みか (ライフが 0 未満) を返す
func (p *ThirtyOnePlayer) IsEliminated() bool { return p.lives < 0 }

// thirtyOnePlayerJSON is the JSON wire format for ThirtyOnePlayer.
type thirtyOnePlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Lives      int         `json:"lv"`
}

// MarshalJSON implements json.Marshaler.
func (p *ThirtyOnePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(thirtyOnePlayerJSON{GamePlayer: p.GamePlayer, Lives: p.lives})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ThirtyOnePlayer) UnmarshalJSON(data []byte) error {
	var j thirtyOnePlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.lives = j.Lives
	return nil
}
