package domain

import "sort"

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
	if len(pp.cards) < 5 {
		pp.handRank = PokerHandHighCard
		return pp.handRank
	}

	values := make([]int, len(pp.cards))
	designs := make([]int, len(pp.cards))
	for i, c := range pp.cards {
		values[i] = c.GetValue()
		designs[i] = c.GetDesign()
	}
	sort.Ints(values)

	// フラッシュチェック
	isFlush := true
	for i := 1; i < len(designs); i++ {
		if designs[i] != designs[0] {
			isFlush = false
			break
		}
	}

	// ストレートチェック
	isStraight := pp.checkStraight(values)

	// カード値の出現回数カウント
	valueCounts := make(map[int]int)
	for _, v := range values {
		valueCounts[v]++
	}
	counts := make([]int, 0)
	for _, c := range valueCounts {
		counts = append(counts, c)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(counts)))

	// ハンドランク判定
	if isFlush && isStraight {
		if pp.checkRoyalStraight(values) {
			pp.handRank = PokerHandRoyalFlush
		} else {
			pp.handRank = PokerHandStraightFlush
		}
	} else if counts[0] == 4 {
		pp.handRank = PokerHandFourOfAKind
	} else if len(counts) >= 2 && counts[0] == 3 && counts[1] == 2 {
		pp.handRank = PokerHandFullHouse
	} else if isFlush {
		pp.handRank = PokerHandFlush
	} else if isStraight {
		pp.handRank = PokerHandStraight
	} else if counts[0] == 3 {
		pp.handRank = PokerHandThreeOfAKind
	} else if len(counts) >= 2 && counts[0] == 2 && counts[1] == 2 {
		pp.handRank = PokerHandTwoPair
	} else if counts[0] == 2 {
		pp.handRank = PokerHandOnePair
	} else {
		pp.handRank = PokerHandHighCard
	}

	return pp.handRank
}

// checkStraight ストレートチェック
func (pp *PokerPlayer) checkStraight(sortedValues []int) bool {
	// 通常のストレート
	isNormal := true
	for i := 1; i < len(sortedValues); i++ {
		if sortedValues[i] != sortedValues[i-1]+1 {
			isNormal = false
			break
		}
	}
	if isNormal {
		return true
	}

	// A-2-3-4-5 のローエーストレート (value: 1,2,3,4,5)
	if len(sortedValues) == 5 &&
		sortedValues[0] == 1 &&
		sortedValues[1] == 2 &&
		sortedValues[2] == 3 &&
		sortedValues[3] == 4 &&
		sortedValues[4] == 5 {
		return true
	}

	// A-10-J-Q-K のハイエーストレート
	if len(sortedValues) == 5 &&
		sortedValues[0] == 1 && sortedValues[1] == 10 &&
		sortedValues[2] == 11 && sortedValues[3] == 12 && sortedValues[4] == 13 {
		return true
	}

	return false
}

// checkRoyalStraight ロイヤルフラッシュかチェック (A-10-J-Q-K)
func (pp *PokerPlayer) checkRoyalStraight(sortedValues []int) bool {
	// sortedValues: [1, 10, 11, 12, 13]
	return len(sortedValues) == 5 &&
		sortedValues[0] == 1 &&
		sortedValues[1] == 10 &&
		sortedValues[2] == 11 &&
		sortedValues[3] == 12 &&
		sortedValues[4] == 13
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
