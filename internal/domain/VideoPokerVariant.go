//go:build !js || !wasm || casino

package domain

// VideoPokerVariantConfig はビデオポーカーのバリアント固有設定を保持する
type VideoPokerVariantConfig struct {
	// Name はバリアント識別子（例: "jacksorbetter", "deuceswild", "jokerpoker"）
	Name string
	// JokerCount はデッキに追加するジョーカー枚数（0 or 1）
	JokerCount int
	// IsWild はカードがワイルドかどうかを判定する関数（nil = ワイルドなし）
	IsWild func(*Card) bool
	// GetResult はハンドを評価し、ランク・配当倍率・ハンド名を返す
	GetResult func(hand []*Card, betAmount int) (rank int, multiplier int, handName string)
}

// --- Jacks or Better ---

// jacksOrBetterPayouts Jacks or Betterのペイアウト倍率テーブル
var jacksOrBetterPayouts = [11]int{
	0,   // HighCard
	1,   // OnePair (Jacks or Better時のみ)
	2,   // TwoPair
	3,   // ThreeOfAKind
	4,   // Straight
	6,   // Flush
	9,   // FullHouse
	25,  // FourOfAKind
	50,  // StraightFlush
	250, // RoyalFlush (5コインベット時は800x)
	0,   // FiveOfAKind (未使用)
}

// JacksOrBetterConfig はJacks or Betterバリアントの設定を返す
func JacksOrBetterConfig() *VideoPokerVariantConfig {
	return &VideoPokerVariantConfig{
		Name:       "jacksorbetter",
		JokerCount: 0,
		IsWild:     nil,
		GetResult:  jacksOrBetterGetResult,
	}
}

// jacksOrBetterGetResult はJacks or Betterのハンド評価・配当計算を行う
func jacksOrBetterGetResult(hand []*Card, betAmount int) (int, int, string) {
	rank := evalFiveCardHand(hand)
	multiplier := jacksOrBetterPayouts[rank]

	// OnePairはJacks or Better条件をチェック
	if rank == PokerHandOnePair {
		if !isJacksOrBetterHand(hand) {
			return rank, 0, ""
		}
		return rank, multiplier, "Jacks or Better"
	}

	// Royal Flushで5コインベット時はボーナス倍率
	if rank == PokerHandRoyalFlush && betAmount == VideoPokerMaxBet {
		return rank, 800, PokerHandNames[rank]
	}

	if multiplier == 0 {
		return rank, 0, ""
	}
	return rank, multiplier, PokerHandNames[rank]
}

// isJacksOrBetterHand はペアがJ以上かどうかを判定する
func isJacksOrBetterHand(hand []*Card) bool {
	valueCounts := make(map[int]int)
	for _, card := range hand {
		valueCounts[card.GetValue()]++
	}
	for value, count := range valueCounts {
		if count >= 2 {
			// A=1, J=11, Q=12, K=13
			if value == 1 || value >= 11 {
				return true
			}
		}
	}
	return false
}

// --- Deuces Wild ---

// DeucesWildConfig はDeuces Wildバリアントの設定を返す
func DeucesWildConfig() *VideoPokerVariantConfig {
	isWild := func(c *Card) bool {
		return c.GetValue() == 2
	}
	return &VideoPokerVariantConfig{
		Name:       "deuceswild",
		JokerCount: 0,
		IsWild:     isWild,
		GetResult: func(hand []*Card, betAmount int) (int, int, string) {
			return deucesWildGetResult(hand, betAmount, isWild)
		},
	}
}

// hasFourDeuces は手札に4枚の2が含まれるかを判定する
func hasFourDeuces(hand []*Card) bool {
	count := 0
	for _, c := range hand {
		if c.GetValue() == 2 {
			count++
		}
	}
	return count == 4
}

// deucesWildGetResult はDeuces Wildのハンド評価・配当計算を行う
func deucesWildGetResult(hand []*Card, betAmount int, isWild func(*Card) bool) (int, int, string) {
	// Four Deucesの特別判定
	if hasFourDeuces(hand) {
		multiplier := 200
		return PokerHandFourOfAKind, multiplier, "Four Deuces"
	}

	bestRank, usedWilds := evalWildHand(hand, isWild)

	// Natural Royal Flush (ワイルド未使用)
	if bestRank == PokerHandRoyalFlush && !usedWilds {
		if betAmount == VideoPokerMaxBet {
			return bestRank, 800, "Natural Royal Flush"
		}
		return bestRank, 250, "Natural Royal Flush"
	}

	// Wild Royal Flush
	if bestRank == PokerHandRoyalFlush && usedWilds {
		return bestRank, 25, "Wild Royal Flush"
	}

	// Deuces Wildペイテーブル
	switch bestRank {
	case PokerHandFiveOfAKind:
		return bestRank, 15, "Five of a Kind"
	case PokerHandStraightFlush:
		return bestRank, 9, "Straight Flush"
	case PokerHandFourOfAKind:
		return bestRank, 5, "Four of a Kind"
	case PokerHandFullHouse:
		return bestRank, 3, "Full House"
	case PokerHandFlush:
		return bestRank, 2, "Flush"
	case PokerHandStraight:
		return bestRank, 2, "Straight"
	case PokerHandThreeOfAKind:
		return bestRank, 1, "Three of a Kind"
	default:
		return bestRank, 0, ""
	}
}

// --- Joker Poker ---

// JokerPokerConfig はJoker Poker (Kings or Better)バリアントの設定を返す
func JokerPokerConfig() *VideoPokerVariantConfig {
	isWild := func(c *Card) bool {
		return c.GetDesign() == CardDesignJoker
	}
	return &VideoPokerVariantConfig{
		Name:       "jokerpoker",
		JokerCount: 1,
		IsWild:     isWild,
		GetResult: func(hand []*Card, betAmount int) (int, int, string) {
			return jokerPokerGetResult(hand, betAmount, isWild)
		},
	}
}

// isKingsOrBetterHand はペアがK以上かどうかを判定する（ジョーカーを除く）
func isKingsOrBetterHand(hand []*Card) bool {
	valueCounts := make(map[int]int)
	for _, card := range hand {
		if card.GetDesign() != CardDesignJoker {
			valueCounts[card.GetValue()]++
		}
	}
	for value, count := range valueCounts {
		if count >= 2 {
			// A=1, K=13
			if value == 1 || value == 13 {
				return true
			}
		}
	}
	return false
}

// jokerPokerGetResult はJoker Pokerのハンド評価・配当計算を行う
func jokerPokerGetResult(hand []*Card, betAmount int, isWild func(*Card) bool) (int, int, string) {
	bestRank, usedWilds := evalWildHand(hand, isWild)

	// Natural Royal Flush (ワイルド未使用)
	if bestRank == PokerHandRoyalFlush && !usedWilds {
		if betAmount == VideoPokerMaxBet {
			return bestRank, 800, "Natural Royal Flush"
		}
		return bestRank, 250, "Natural Royal Flush"
	}

	// Wild Royal Flush
	if bestRank == PokerHandRoyalFlush && usedWilds {
		return bestRank, 100, "Wild Royal Flush"
	}

	// Joker Pokerペイテーブル
	switch bestRank {
	case PokerHandFiveOfAKind:
		return bestRank, 200, "Five of a Kind"
	case PokerHandStraightFlush:
		return bestRank, 50, "Straight Flush"
	case PokerHandFourOfAKind:
		return bestRank, 20, "Four of a Kind"
	case PokerHandFullHouse:
		return bestRank, 7, "Full House"
	case PokerHandFlush:
		return bestRank, 5, "Flush"
	case PokerHandStraight:
		return bestRank, 3, "Straight"
	case PokerHandThreeOfAKind:
		return bestRank, 2, "Three of a Kind"
	case PokerHandTwoPair:
		return bestRank, 1, "Two Pair"
	case PokerHandOnePair:
		if isKingsOrBetterHand(hand) {
			return bestRank, 1, "Kings or Better"
		}
		return bestRank, 0, ""
	default:
		return bestRank, 0, ""
	}
}
