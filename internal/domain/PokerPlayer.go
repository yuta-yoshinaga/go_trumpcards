package domain

// ポーカーハンドランク定数
const (
	PokerHandHighCard      = 0
	PokerHandOnePair       = 1
	PokerHandTwoPair       = 2
	PokerHandThreeOfAKind  = 3
	PokerHandStraight      = 4
	PokerHandFlush         = 5
	PokerHandFullHouse     = 6
	PokerHandFourOfAKind   = 7
	PokerHandStraightFlush = 8
	PokerHandRoyalFlush    = 9
)

// PokerHandNames ポーカーハンド名
var PokerHandNames = []string{
	"High Card",
	"One Pair",
	"Two Pair",
	"Three of a Kind",
	"Straight",
	"Flush",
	"Full House",
	"Four of a Kind",
	"Straight Flush",
	"Royal Flush",
}

// PokerPlayer ポーカープレイヤークラス
type PokerPlayer struct {
	Player       // 親クラス
	handRank int // ハンドランク
	chips    int // チップ
}

// NewPokerPlayer コンストラクタ
func NewPokerPlayer() *PokerPlayer {
	return &PokerPlayer{
		Player: Player{
			cards: make([]*Card, 0),
		},
		handRank: 0,
		chips:    0,
	}
}

// AddCard カード追加
func (pp *PokerPlayer) AddCard(card *Card) {
	pp.cards = append(pp.cards, card)
}

// ExchangeCard 指定インデックスのカードを交換
func (pp *PokerPlayer) ExchangeCard(idx int, card *Card) {
	if 0 <= idx && idx < len(pp.cards) {
		pp.cards[idx] = card
	}
}

// EvalHand ハンド評価
func (pp *PokerPlayer) EvalHand() int {
	pp.handRank = evalFiveCardHand(pp.cards)
	return pp.handRank
}

// GetHandRank ハンドランク取得
func (pp *PokerPlayer) GetHandRank() int {
	return pp.handRank
}

// GetHandName ハンド名取得
func (pp *PokerPlayer) GetHandName() string {
	if 0 <= pp.handRank && pp.handRank < len(PokerHandNames) {
		return PokerHandNames[pp.handRank]
	}
	return "Unknown"
}

// GetChips チップ取得
func (pp *PokerPlayer) GetChips() int {
	return pp.chips
}

// SetChips チップ設定
func (pp *PokerPlayer) SetChips(chips int) {
	pp.chips = chips
}

// AddChips チップ追加
func (pp *PokerPlayer) AddChips(amount int) {
	pp.chips += amount
}

// SubtractChips チップ減算 (不足時はfalseを返す)
func (pp *PokerPlayer) SubtractChips(amount int) bool {
	if pp.chips < amount {
		return false
	}
	pp.chips -= amount
	return true
}
