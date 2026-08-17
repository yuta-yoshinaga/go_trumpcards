//go:build !js || !wasm || casino

package domain

import "encoding/json"

// BlackJackBustOver はこれを超えるとバーストになる合計 (21)。
//
// 数字そのものは規則。表示側が「21 を超えたらバースト」と書き写すと、
// 得点の計算と案内が別々の 21 を持つことになる。
const BlackJackBustOver = 21

// BlackJackHand ブラックジャックハンド（分割対応）
type BlackJackHand struct {
	cards       []*Card
	bet         int
	stood       bool
	doubled     bool
	busted      bool
	surrendered bool
	fromSplit   bool
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

// SetCard 指定番目のカードを差し替える（範囲外なら no-op）。
func (h *BlackJackHand) SetCard(idx int, card *Card) {
	if 0 <= idx && idx < len(h.cards) {
		h.cards[idx] = card
	}
}

// GetCardsSize カード枚数取得
func (h *BlackJackHand) GetCardsSize() int {
	return len(h.cards)
}

// calcScore カードスライスからスコアとソフト状態を計算する内部ヘルパー
// isSoft=true はエースが11として有効に働いていることを意味する
func calcScore(cards []*Card) (score int, isSoft bool) {
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
	for score > BlackJackBustOver && aceCount > 0 {
		score -= 10
		aceCount--
	}
	return score, aceCount > 0
}

// CalculateBlackJackScore カードスライスからブラックジャックスコアを計算する共通関数
func CalculateBlackJackScore(cards []*Card) int {
	score, _ := calcScore(cards)
	return score
}

// GetScore 手札から現在のスコア計算
func (h *BlackJackHand) GetScore() int {
	return CalculateBlackJackScore(h.cards)
}

// IsSoft ソフトハンド（11として有効なエースを含む）かどうか判定
func (h *BlackJackHand) IsSoft() bool {
	_, isSoft := calcScore(h.cards)
	return isSoft
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

// IsSurrendered サレンダー済みか
func (h *BlackJackHand) IsSurrendered() bool {
	return h.surrendered
}

// SetSurrendered サレンダー状態設定
func (h *BlackJackHand) SetSurrendered(surrendered bool) {
	h.surrendered = surrendered
}

// IsFromSplit スプリットから生じたハンドか
func (h *BlackJackHand) IsFromSplit() bool {
	return h.fromSplit
}

// SetFromSplit スプリット由来フラグ設定
func (h *BlackJackHand) SetFromSplit(fromSplit bool) {
	h.fromSplit = fromSplit
}

// CanSurrender サレンダー可能か（2枚でスタンド/バースト/サレンダー前）
func (h *BlackJackHand) CanSurrender() bool {
	return len(h.cards) == 2 && !h.stood && !h.busted && !h.surrendered
}

// IsFinished ハンドが完了しているか
func (h *BlackJackHand) IsFinished() bool {
	return h.stood || h.busted || h.surrendered
}

// Reset ハンドリセット
func (h *BlackJackHand) Reset() {
	h.cards = make([]*Card, 0)
	h.bet = 0
	h.stood = false
	h.doubled = false
	h.busted = false
	h.surrendered = false
	h.fromSplit = false
}

// blackJackHandJSON is the JSON wire format for BlackJackHand.
type blackJackHandJSON struct {
	Cards       []*Card `json:"c"`
	Bet         int     `json:"b"`
	Stood       bool    `json:"st"`
	Doubled     bool    `json:"db"`
	Busted      bool    `json:"bu"`
	Surrendered bool    `json:"sr"`
	FromSplit   bool    `json:"fs"`
}

// MarshalJSON implements json.Marshaler.
func (h *BlackJackHand) MarshalJSON() ([]byte, error) {
	return json.Marshal(blackJackHandJSON{
		Cards:       h.cards,
		Bet:         h.bet,
		Stood:       h.stood,
		Doubled:     h.doubled,
		Busted:      h.busted,
		Surrendered: h.surrendered,
		FromSplit:   h.fromSplit,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (h *BlackJackHand) UnmarshalJSON(data []byte) error {
	var j blackJackHandJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	h.cards = j.Cards
	if h.cards == nil {
		h.cards = make([]*Card, 0)
	}
	h.bet = j.Bet
	h.stood = j.Stood
	h.doubled = j.Doubled
	h.busted = j.Busted
	h.surrendered = j.Surrendered
	h.fromSplit = j.FromSplit
	return nil
}
