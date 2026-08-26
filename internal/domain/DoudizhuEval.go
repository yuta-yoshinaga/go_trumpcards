//go:build !js || !wasm || extra4

package domain

import "sort"

// 斗地主カード強さ定数
const (
	doudizhuStrengthSmallJoker = 16
	doudizhuStrengthBigJoker   = 17
)

// DoudizhuComboType 斗地主の役タイプ
type DoudizhuComboType int

const (
	// DoudizhuComboPass パス
	DoudizhuComboPass DoudizhuComboType = iota
	// DoudizhuComboSingle 単張
	DoudizhuComboSingle
	// DoudizhuComboPair 対子
	DoudizhuComboPair
	// DoudizhuComboTrio 三条
	DoudizhuComboTrio
	// DoudizhuComboTrioSingle 三帯一
	DoudizhuComboTrioSingle
	// DoudizhuComboTrioPair 三帯二
	DoudizhuComboTrioPair
	// DoudizhuComboStraight 順子 (5枚以上)
	DoudizhuComboStraight
	// DoudizhuComboConsecutivePair 連対 (3組以上)
	DoudizhuComboConsecutivePair
	// DoudizhuComboAirplane 飛機 (連続三条)
	DoudizhuComboAirplane
	// DoudizhuComboAirplaneSingle 飛機帯単 (連続三条+単張翼)
	DoudizhuComboAirplaneSingle
	// DoudizhuComboAirplanePair 飛機帯対 (連続三条+対子翼)
	DoudizhuComboAirplanePair
	// DoudizhuComboBomb 炸弾 (同数4枚)
	DoudizhuComboBomb
	// DoudizhuComboRocket 火箭 (ジョーカー2枚)
	DoudizhuComboRocket
)

// DoudizhuCombo 斗地主の役
type DoudizhuCombo struct {
	Type   DoudizhuComboType
	Cards  []*Card
	Rank   int // 比較用の主ランク (三条の値など)
	Length int // 連続数 (順子/連対/飛機のチェーン長)
}

// DoudizhuCardStrength カードの強さを返す
// 3(3) < 4(4) < ... < K(13) < A(14) < 2(15) < SmallJoker(16) < BigJoker(17)
func DoudizhuCardStrength(card *Card) int {
	if card.GetDesign() == CardDesignJoker {
		if card.GetValue() >= 2 {
			return doudizhuStrengthBigJoker
		}
		return doudizhuStrengthSmallJoker
	}
	v := card.GetValue()
	if v == 1 {
		return 14
	}
	if v == 2 {
		return 15
	}
	return v
}

// IsBigJoker 大ジョーカー判定
func IsBigJoker(card *Card) bool {
	return card.GetDesign() == CardDesignJoker && card.GetValue() >= 2
}

// IsSmallJoker 小ジョーカー判定
func IsSmallJoker(card *Card) bool {
	return card.GetDesign() == CardDesignJoker && card.GetValue() == 1
}

// DoudizhuClassifyCombo カード配列から役を判定する (無効ならnil)
func DoudizhuClassifyCombo(cards []*Card) *DoudizhuCombo {
	n := len(cards)
	if n == 0 {
		return nil
	}

	if n == 2 && isRocket(cards) {
		return &DoudizhuCombo{Type: DoudizhuComboRocket, Cards: cards, Rank: doudizhuStrengthBigJoker, Length: 1}
	}

	ranks := doudizhuRankCounts(cards)

	if n == 1 {
		return &DoudizhuCombo{Type: DoudizhuComboSingle, Cards: cards, Rank: DoudizhuCardStrength(cards[0]), Length: 1}
	}
	if n == 2 {
		if ranks[0].count == 2 {
			return &DoudizhuCombo{Type: DoudizhuComboPair, Cards: cards, Rank: ranks[0].strength, Length: 1}
		}
		return nil
	}
	if n == 3 && ranks[0].count == 3 {
		return &DoudizhuCombo{Type: DoudizhuComboTrio, Cards: cards, Rank: ranks[0].strength, Length: 1}
	}
	if n == 4 {
		if len(ranks) == 1 && ranks[0].count == 4 {
			return &DoudizhuCombo{Type: DoudizhuComboBomb, Cards: cards, Rank: ranks[0].strength, Length: 1}
		}
		if combo := classifyTrioKicker(ranks, cards, n); combo != nil {
			return combo
		}
	}
	if n == 5 {
		if combo := classifyTrioKicker(ranks, cards, n); combo != nil {
			return combo
		}
	}

	if combo := classifyStraight(ranks, cards, n); combo != nil {
		return combo
	}
	if combo := classifyConsecutivePair(ranks, cards, n); combo != nil {
		return combo
	}
	if combo := classifyAirplane(ranks, cards, n); combo != nil {
		return combo
	}

	return nil
}

// DoudizhuCanBeat play が table に勝てるか判定
func DoudizhuCanBeat(play, table *DoudizhuCombo) bool {
	if play.Type == DoudizhuComboRocket {
		return true
	}
	if play.Type == DoudizhuComboBomb {
		if table.Type == DoudizhuComboBomb {
			return play.Rank > table.Rank
		}
		if table.Type == DoudizhuComboRocket {
			return false
		}
		return true
	}
	if play.Type != table.Type {
		return false
	}
	if play.Length != table.Length {
		return false
	}
	return play.Rank > table.Rank
}

// --- internal helpers ---

type rankCount struct {
	strength int
	count    int
}

func doudizhuRankCounts(cards []*Card) []rankCount {
	freq := make(map[int]int)
	for _, c := range cards {
		s := DoudizhuCardStrength(c)
		freq[s]++
	}
	result := make([]rankCount, 0, len(freq))
	for s, cnt := range freq {
		result = append(result, rankCount{strength: s, count: cnt})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].count != result[j].count {
			return result[i].count > result[j].count
		}
		return result[i].strength < result[j].strength
	})
	return result
}

func isRocket(cards []*Card) bool {
	if len(cards) != 2 {
		return false
	}
	return IsJoker(cards[0]) && IsJoker(cards[1])
}

func classifyTrioKicker(ranks []rankCount, cards []*Card, n int) *DoudizhuCombo {
	if len(ranks) < 1 {
		return nil
	}
	trioRank := -1
	for _, r := range ranks {
		if r.count == 3 {
			trioRank = r.strength
			break
		}
	}
	if trioRank < 0 {
		return nil
	}

	kickerCount := n - 3
	if kickerCount == 1 {
		return &DoudizhuCombo{Type: DoudizhuComboTrioSingle, Cards: cards, Rank: trioRank, Length: 1}
	}
	if kickerCount == 2 {
		for _, r := range ranks {
			if r.strength != trioRank && r.count == 2 {
				return &DoudizhuCombo{Type: DoudizhuComboTrioPair, Cards: cards, Rank: trioRank, Length: 1}
			}
		}
	}
	return nil
}

func isChainable(strength int) bool {
	return strength >= 3 && strength <= 14
}

func classifyStraight(ranks []rankCount, cards []*Card, n int) *DoudizhuCombo {
	if n < 5 {
		return nil
	}
	for _, r := range ranks {
		if r.count != 1 {
			return nil
		}
		if !isChainable(r.strength) {
			return nil
		}
	}
	strengths := make([]int, 0, len(ranks))
	for _, r := range ranks {
		strengths = append(strengths, r.strength)
	}
	sort.Ints(strengths)
	if len(strengths) != n {
		return nil
	}
	for i := 1; i < len(strengths); i++ {
		if strengths[i]-strengths[i-1] != 1 {
			return nil
		}
	}
	return &DoudizhuCombo{Type: DoudizhuComboStraight, Cards: cards, Rank: strengths[0], Length: n}
}

func classifyConsecutivePair(ranks []rankCount, cards []*Card, n int) *DoudizhuCombo {
	if n < 6 || n%2 != 0 {
		return nil
	}
	pairCount := n / 2
	if pairCount < 3 {
		return nil
	}
	for _, r := range ranks {
		if r.count != 2 {
			return nil
		}
		if !isChainable(r.strength) {
			return nil
		}
	}
	strengths := make([]int, 0, len(ranks))
	for _, r := range ranks {
		strengths = append(strengths, r.strength)
	}
	sort.Ints(strengths)
	if len(strengths) != pairCount {
		return nil
	}
	for i := 1; i < len(strengths); i++ {
		if strengths[i]-strengths[i-1] != 1 {
			return nil
		}
	}
	return &DoudizhuCombo{Type: DoudizhuComboConsecutivePair, Cards: cards, Rank: strengths[0], Length: pairCount}
}

func classifyAirplane(ranks []rankCount, cards []*Card, n int) *DoudizhuCombo {
	trios := make([]int, 0)
	for _, r := range ranks {
		if r.count == 3 && isChainable(r.strength) {
			trios = append(trios, r.strength)
		}
	}
	if len(trios) < 2 {
		return nil
	}
	sort.Ints(trios)

	chain := findLongestConsecutive(trios)
	if len(chain) < 2 {
		return nil
	}
	chainLen := len(chain)

	// Pure airplane: exactly chainLen consecutive trios, no wings.
	if n == chainLen*3 {
		return &DoudizhuCombo{Type: DoudizhuComboAirplane, Cards: cards, Rank: chain[0], Length: chainLen}
	}

	// Airplane + single wings: chainLen trios + chainLen single kickers.
	if n == chainLen*4 {
		return &DoudizhuCombo{Type: DoudizhuComboAirplaneSingle, Cards: cards, Rank: chain[0], Length: chainLen}
	}

	// Airplane + pair wings: chainLen trios + chainLen pair kickers. Every
	// rank's leftover count (after removing the chain trios) must be even so
	// the kickers partition cleanly into pairs.
	if n == chainLen*5 {
		inChain := func(strength int) bool {
			for _, c := range chain {
				if strength == c {
					return true
				}
			}
			return false
		}
		for _, r := range ranks {
			remCount := r.count
			if inChain(r.strength) {
				remCount -= 3
			}
			if remCount%2 != 0 {
				return nil
			}
		}
		return &DoudizhuCombo{Type: DoudizhuComboAirplanePair, Cards: cards, Rank: chain[0], Length: chainLen}
	}

	return nil
}

func findLongestConsecutive(sorted []int) []int {
	if len(sorted) == 0 {
		return nil
	}
	best := []int{sorted[0]}
	current := []int{sorted[0]}
	for i := 1; i < len(sorted); i++ {
		if sorted[i] == sorted[i-1]+1 {
			current = append(current, sorted[i])
		} else {
			current = []int{sorted[i]}
		}
		if len(current) > len(best) {
			best = make([]int, len(current))
			copy(best, current)
		}
	}
	return best
}
