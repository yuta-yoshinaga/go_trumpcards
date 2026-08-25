//go:build !js || !wasm || extra2

package domain

// ContinentalRummyPlayer はコンチネンタル・ラミーの 1 席。
type ContinentalRummyPlayer struct {
	// isHuman は人間の席か。
	isHuman bool
	// hand は手札。
	hand []*Card
	// score は通算得点。**勝った側が集める方式**なので、増える一方。
	score int
	// melds は上がったときに並べたシーケンス。上がっていなければ空。
	melds [][]*Card
}

// NewContinentalRummyPlayer はコンストラクタ。
func NewContinentalRummyPlayer(isHuman bool) *ContinentalRummyPlayer {
	return &ContinentalRummyPlayer{isHuman: isHuman, hand: make([]*Card, 0, ContinentalRummyHandSize+1)}
}

// GetIsHuman は人間の席かを返す。
func (p *ContinentalRummyPlayer) GetIsHuman() bool { return p.isHuman }

// GetHand は手札を返す。
func (p *ContinentalRummyPlayer) GetHand() []*Card { return p.hand }

// GetCardsSize は手札の枚数を返す。
func (p *ContinentalRummyPlayer) GetCardsSize() int { return len(p.hand) }

// GetCard は i 番目の手札を返す。範囲外なら nil。
func (p *ContinentalRummyPlayer) GetCard(i int) *Card {
	if i < 0 || i >= len(p.hand) {
		return nil
	}
	return p.hand[i]
}

// AddCard は手札に 1 枚加える。
func (p *ContinentalRummyPlayer) AddCard(c *Card) {
	if c != nil {
		p.hand = append(p.hand, c)
	}
}

// RemoveCard は i 番目の手札を抜いて返す。範囲外なら nil。
func (p *ContinentalRummyPlayer) RemoveCard(i int) *Card {
	if i < 0 || i >= len(p.hand) {
		return nil
	}
	c := p.hand[i]
	p.hand = append(p.hand[:i], p.hand[i+1:]...)
	return c
}

// GetScore は通算得点を返す。
func (p *ContinentalRummyPlayer) GetScore() int { return p.score }

// AddScore は得点を加える。
func (p *ContinentalRummyPlayer) AddScore(n int) { p.score += n }

// GetMelds は並べたシーケンスを返す。
func (p *ContinentalRummyPlayer) GetMelds() [][]*Card { return p.melds }

// SetMelds は並べたシーケンスを記録する。
func (p *ContinentalRummyPlayer) SetMelds(m [][]*Card) { p.melds = m }

// ResetRound はラウンドぶんの状態を消す。**得点は残す。**
func (p *ContinentalRummyPlayer) ResetRound() {
	p.hand = p.hand[:0]
	p.melds = nil
}

// ClearHand は手札を空にする。上がって場に並べたときに使う。
func (p *ContinentalRummyPlayer) ClearHand() { p.hand = p.hand[:0] }
