package domain

import "encoding/json"

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

// blackJackCpuSeatJSON is the JSON wire format for BlackJackCpuSeat.
type blackJackCpuSeatJSON struct {
	Player       *BlackJackPlayer `json:"p"`
	Hands        []*BlackJackHand `json:"hs"`
	InsuranceBet int              `json:"ib"`
}

// MarshalJSON implements json.Marshaler.
func (c *BlackJackCpuSeat) MarshalJSON() ([]byte, error) {
	return json.Marshal(blackJackCpuSeatJSON{
		Player:       c.player,
		Hands:        c.hands,
		InsuranceBet: c.insuranceBet,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *BlackJackCpuSeat) UnmarshalJSON(data []byte) error {
	var j blackJackCpuSeatJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.player = j.Player
	if c.player == nil {
		c.player = NewBlackJackPlayer()
	}
	c.hands = j.Hands
	if c.hands == nil {
		c.hands = []*BlackJackHand{NewBlackJackHand()}
	}
	c.insuranceBet = j.InsuranceBet
	return nil
}
