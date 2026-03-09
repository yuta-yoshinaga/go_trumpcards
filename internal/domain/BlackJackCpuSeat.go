package domain

// BlackJackCpuSeat CPUプレイヤー席
type BlackJackCpuSeat struct {
	player       *BlackJackPlayer
	hands        []*BlackJackHand
	insuranceBet int
}

// NewBlackJackCpuSeat コンストラクタ
func NewBlackJackCpuSeat() *BlackJackCpuSeat {
	p := NewBlackJackPlayer()
	p.SetChips(BJDefaultChips)
	return &BlackJackCpuSeat{
		player: p,
		hands:  []*BlackJackHand{NewBlackJackHand()},
	}
}

// GetPlayer プレイヤー取得
func (c *BlackJackCpuSeat) GetPlayer() *BlackJackPlayer {
	return c.player
}

// GetHands ハンド一覧取得
func (c *BlackJackCpuSeat) GetHands() []*BlackJackHand {
	return c.hands
}

// SetHands ハンド設定（スプリット用）
func (c *BlackJackCpuSeat) SetHands(hands []*BlackJackHand) {
	c.hands = hands
}

// GetInsuranceBet インシュランスベット取得
func (c *BlackJackCpuSeat) GetInsuranceBet() int {
	return c.insuranceBet
}

// SetInsuranceBet インシュランスベット設定
func (c *BlackJackCpuSeat) SetInsuranceBet(bet int) {
	c.insuranceBet = bet
}

// Reset ハンドリセット（チップは保持）
func (c *BlackJackCpuSeat) Reset() {
	c.player.Reset()
	c.hands = []*BlackJackHand{NewBlackJackHand()}
	c.insuranceBet = 0
}
