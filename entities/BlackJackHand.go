package entities

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

// GetScore 手札から現在のスコア計算（BlackJackPlayerと同じロジック）
func (h *BlackJackHand) GetScore() int {
	aceFlag := false
	score := 0
	for _, card := range h.cards {
		value := card.GetValue()
		if 2 <= value && value <= 10 {
			score += value
		} else if 11 <= value && value <= 13 {
			score += 10
		} else {
			if aceFlag {
				score++
			} else {
				aceFlag = true
			}
		}
	}
	if aceFlag {
		if score+11 <= 21 {
			score += 11
		} else {
			score++
		}
	}
	return score
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
