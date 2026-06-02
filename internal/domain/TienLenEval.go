package domain

import "sort"

// Tien Len card value strength: 3 < 4 < ... < K < A < 2
// Suit strength: ♠ < ♣ < ♦ < ♥ (spades lowest, hearts highest)
// Combined strength = valueStrength * 4 + suitStrength
// (example: ♠2 is weaker than ♥2)

// tienLenValueStrength returns the value strength (3=0, 4=1, ..., K=10, A=11, 2=12).
func tienLenValueStrength(v int) int {
	if v == 2 {
		return 12
	}
	if v == 1 {
		return 11
	}
	return v - 3
}

// tienLenSuitStrength returns the suit strength (♠=0, ♣=1, ♦=2, ♥=3).
func tienLenSuitStrength(design int) int {
	switch design {
	case CardDesignSpade:
		return 0
	case CardDesignClover:
		return 1
	case CardDesignDiamond:
		return 2
	case CardDesignHeart:
		return 3
	default:
		return 0
	}
}

// TienLenCardStrength returns the combined strength of a card.
func TienLenCardStrength(card *Card) int {
	return tienLenValueStrength(card.GetValue())*4 + tienLenSuitStrength(card.GetDesign())
}

// TienLenPlayType プレイの種類
type TienLenPlayType int

// TienLenPlayType定数
const (
	TienLenPlayInvalid      TienLenPlayType = 0
	TienLenPlaySingle       TienLenPlayType = 1
	TienLenPlayPair         TienLenPlayType = 2
	TienLenPlayTriple       TienLenPlayType = 3
	TienLenPlayStraight     TienLenPlayType = 4 // 3枚以上の連続(スート混在可)
	TienLenPlayThreePairRun TienLenPlayType = 5 // 連続する3つのペア(チョップ役)
	TienLenPlayFourOfAKind  TienLenPlayType = 6 // フォーカード(チョップ役)
)

// tienLenAllSameValue は全カードが同じランクかを返す。
func tienLenAllSameValue(cards []*Card) bool {
	for i := 1; i < len(cards); i++ {
		if cards[i].GetValue() != cards[0].GetValue() {
			return false
		}
	}
	return len(cards) > 0
}

// tienLenClassifyPlay classifies a set of cards into its Tien Len play type.
func tienLenClassifyPlay(cards []*Card) TienLenPlayType {
	n := len(cards)
	switch n {
	case 0:
		return TienLenPlayInvalid
	case 1:
		return TienLenPlaySingle
	case 2:
		if tienLenAllSameValue(cards) {
			return TienLenPlayPair
		}
		return TienLenPlayInvalid
	case 3:
		if tienLenAllSameValue(cards) {
			return TienLenPlayTriple
		}
		if tienLenCheckStraight(cards) {
			return TienLenPlayStraight
		}
		return TienLenPlayInvalid
	case 4:
		if tienLenAllSameValue(cards) {
			return TienLenPlayFourOfAKind
		}
		if tienLenCheckStraight(cards) {
			return TienLenPlayStraight
		}
		return TienLenPlayInvalid
	case 6:
		if tienLenCheckThreePairRun(cards) {
			return TienLenPlayThreePairRun
		}
		if tienLenCheckStraight(cards) {
			return TienLenPlayStraight
		}
		return TienLenPlayInvalid
	default:
		if tienLenCheckStraight(cards) {
			return TienLenPlayStraight
		}
		return TienLenPlayInvalid
	}
}

// tienLenCheckStraight reports whether the cards form a straight (run) of length >= 3.
// In Tien Len, 2 (the highest single) cannot be part of a straight, and suits may be mixed.
func tienLenCheckStraight(cards []*Card) bool {
	n := len(cards)
	if n < 3 {
		return false
	}
	strs := make([]int, n)
	for i, c := range cards {
		if c.GetValue() == 2 {
			return false
		}
		strs[i] = tienLenValueStrength(c.GetValue())
	}
	sort.Ints(strs)
	for i := 1; i < n; i++ {
		if strs[i] != strs[i-1]+1 {
			return false
		}
	}
	return true
}

// tienLenCheckThreePairRun reports whether the cards form three consecutive pairs
// (e.g. 4-4-5-5-6-6). 2 cannot appear in a chop.
func tienLenCheckThreePairRun(cards []*Card) bool {
	if len(cards) != 6 {
		return false
	}
	freq := make(map[int]int)
	for _, c := range cards {
		if c.GetValue() == 2 {
			return false
		}
		freq[c.GetValue()]++
	}
	if len(freq) != 3 {
		return false
	}
	strs := make([]int, 0, 3)
	for v, cnt := range freq {
		if cnt != 2 {
			return false
		}
		strs = append(strs, tienLenValueStrength(v))
	}
	sort.Ints(strs)
	return strs[1] == strs[0]+1 && strs[2] == strs[1]+1
}

// tienLenPlayStrength returns a comparable strength value for a play.
// Higher is stronger. Returns -1 for invalid plays.
func tienLenPlayStrength(cards []*Card, playType TienLenPlayType) int {
	switch playType {
	case TienLenPlaySingle:
		return TienLenCardStrength(cards[0])
	// A straight's cards all have distinct values, so the highest TienLenCardStrength
	// always belongs to the highest-value card (value dominates suit). It therefore
	// shares the highest-card comparison used by pairs/triples/three-pair-runs.
	case TienLenPlayPair, TienLenPlayTriple, TienLenPlayThreePairRun, TienLenPlayStraight:
		return tienLenGroupStrength(cards)
	case TienLenPlayFourOfAKind:
		return tienLenValueStrength(cards[0].GetValue())
	default:
		return -1
	}
}

// tienLenGroupStrength returns strength by the highest card (value then suit).
func tienLenGroupStrength(cards []*Card) int {
	maxStr := 0
	for _, c := range cards {
		if s := TienLenCardStrength(c); s > maxStr {
			maxStr = s
		}
	}
	return maxStr
}

// tienLenIsBomb reports whether a play type is a chop (bomb).
func tienLenIsBomb(playType TienLenPlayType) bool {
	return playType == TienLenPlayThreePairRun || playType == TienLenPlayFourOfAKind
}

// tienLenIsSingleTwo reports whether the table holds a single 2 (the "pig").
func tienLenIsSingleTwo(tableCards []*Card, tablePlayType TienLenPlayType) bool {
	return tablePlayType == TienLenPlaySingle && len(tableCards) == 1 && tableCards[0].GetValue() == 2
}

// tienLenIsPlayable checks if cards can be played on the table.
func tienLenIsPlayable(cards []*Card, tableCards []*Card, tablePlayType TienLenPlayType) bool {
	playType := tienLenClassifyPlay(cards)
	if playType == TienLenPlayInvalid {
		return false
	}

	// Leading: any valid combination may be played.
	if tableCards == nil {
		return true
	}

	// Chops (bombs) may cut a single 2 and weaker chops regardless of table type.
	if tienLenIsBomb(playType) {
		if tienLenIsSingleTwo(tableCards, tablePlayType) {
			return true
		}
		switch tablePlayType {
		case TienLenPlayFourOfAKind:
			return playType == TienLenPlayFourOfAKind &&
				tienLenPlayStrength(cards, playType) > tienLenPlayStrength(tableCards, tablePlayType)
		case TienLenPlayThreePairRun:
			if playType == TienLenPlayFourOfAKind {
				return true
			}
			return tienLenPlayStrength(cards, playType) > tienLenPlayStrength(tableCards, tablePlayType)
		default:
			// A chop cannot be played onto an unrelated combination (e.g. a straight).
			return false
		}
	}

	// Normal play: must match the table's type and card count, and be stronger.
	if playType != tablePlayType || len(cards) != len(tableCards) {
		return false
	}
	return tienLenPlayStrength(cards, playType) > tienLenPlayStrength(tableCards, tablePlayType)
}
