package domain

// BlackJackHand ブラックジャックハンド（分割対応）
type BlackJackHand struct {
	cards   []*Card
	bet     int
	stood   bool
	doubled bool
	busted  bool
}

// NewBlackJackHand コンストラクタ
func NewBlackJackHand() *BlackJackHand {
	return &BlackJackHand{
		cards: make([]*Card, 0),
	}
}

// AddCard カード追加
func (h *BlackJackHand) AddCard(card *Card) {
	h.cards = append(h.cards, card)
}

// GetCards カード一覧取得
func (h *BlackJackHand) GetCards() []*Card {
	return h.cards
}

// GetCard 指定番目のカード取得
func (h *BlackJackHand) GetCard(idx int) *Card {
	if 0 <= idx && idx < len(h.cards) {
		return h.cards[idx]
	}
	return nil
}

// GetCardsSize カード枚数取得
func (h *BlackJackHand) GetCardsSize() int {
	return len(h.cards)
}

// CalculateBlackJackScore カードスライスからブラックジャックスコアを計算する共通関数
func CalculateBlackJackScore(cards []*Card) int {
	score := 0
	aceCount := 0
	for _, card := range cards {
		value := card.GetValue()
		if value == 1 {
			aceCount++
			score += 11
		} else if value >= 10 {
			score += 10
		} else {
			score += value
		}
	}
	// スコアが21を超えている場合、エースを1として再計算
	for score > 21 && aceCount > 0 {
		score -= 10
		aceCount--
	}
	return score
}

// GetScore 手札から現在のスコア計算
func (h *BlackJackHand) GetScore() int {
	return CalculateBlackJackScore(h.cards)
}

// GetBet ベット額取得
func (h *BlackJackHand) GetBet() int {
	return h.bet
}

// SetBet ベット額設定
func (h *BlackJackHand) SetBet(bet int) {
	h.bet = bet
}

// IsStood スタンド済みか
func (h *BlackJackHand) IsStood() bool {
	return h.stood
}

// SetStood スタンド状態設定
func (h *BlackJackHand) SetStood(stood bool) {
	h.stood = stood
}

// IsDoubled ダブルダウン済みか
func (h *BlackJackHand) IsDoubled() bool {
	return h.doubled
}

// SetDoubled ダブルダウン状態設定
func (h *BlackJackHand) SetDoubled(doubled bool) {
	h.doubled = doubled
}

// IsBusted バーストしたか
func (h *BlackJackHand) IsBusted() bool {
	return h.busted
}

// SetBusted バースト状態設定
func (h *BlackJackHand) SetBusted(busted bool) {
	h.busted = busted
}

// IsBlackJack ナチュラルブラックジャックか（2枚でスコア21）
func (h *BlackJackHand) IsBlackJack() bool {
	return len(h.cards) == 2 && h.GetScore() == 21
}

// bjValue ブラックジャックにおけるカードの値（Face cards = 10）
func bjValue(card *Card) int {
	v := card.GetValue()
	if v >= 11 {
		return 10
	}
	return v
}

// CanSplit スプリット可能か（2枚でBJ値が同じ）
func (h *BlackJackHand) CanSplit() bool {
	if len(h.cards) != 2 {
		return false
	}
	return bjValue(h.cards[0]) == bjValue(h.cards[1])
}

// IsFinished ハンドが完了しているか
func (h *BlackJackHand) IsFinished() bool {
	return h.stood || h.busted
}

// Reset ハンドリセット
func (h *BlackJackHand) Reset() {
	h.cards = make([]*Card, 0)
	h.bet = 0
	h.stood = false
	h.doubled = false
	h.busted = false
}
