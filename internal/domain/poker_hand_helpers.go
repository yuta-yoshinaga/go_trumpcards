// Package domain ポーカーハンド共通ヘルパー。
//
// No build tag: these helpers are compiled into every build, including all
// Cloudflare Worker WASM binaries, so games can share them regardless of which
// category worker they are assigned to.
package domain

import "sort"

// combinations n枚からk枚を選ぶ全組み合わせを返す
func combinations(cards []*Card, k int) [][]*Card {
	var result [][]*Card
	n := len(cards)
	if k > n {
		return result
	}
	combo := make([]int, k)
	var generate func(start, idx int)
	generate = func(start, idx int) {
		if idx == k {
			hand := make([]*Card, k)
			for i, ci := range combo {
				hand[i] = cards[ci]
			}
			result = append(result, hand)
			return
		}
		for i := start; i <= n-(k-idx); i++ {
			combo[idx] = i
			generate(i+1, idx+1)
		}
	}
	generate(0, 0)
	return result
}

// isWheelHand ホイール (A-2-3-4-5) かどうか判定
func isWheelHand(cards []*Card) bool {
	if len(cards) != 5 {
		return false
	}
	vals := make([]int, 5)
	for i, c := range cards {
		vals[i] = c.GetValue()
	}
	sort.Ints(vals)
	return vals[0] == 1 && vals[1] == 2 && vals[2] == 3 && vals[3] == 4 && vals[4] == 5
}

// tieBreakValues カード値リストを (出現回数 DESC, 値 DESC) でソートしたユニーク値リストを返す
// ペア系ハンドの正しいタイブレーク順序を保証する
func tieBreakValues(vals []int) []int {
	freq := make(map[int]int)
	for _, v := range vals {
		freq[v]++
	}
	unique := make([]int, 0, len(freq))
	for v := range freq {
		unique = append(unique, v)
	}
	sort.Slice(unique, func(i, j int) bool {
		if freq[unique[i]] != freq[unique[j]] {
			return freq[unique[i]] > freq[unique[j]]
		}
		return unique[i] > unique[j]
	})
	return unique
}

// compareHighCardsSlice 2つの5枚ハンドのハイカード比較 (a > b: 1, a < b: -1, a == b: 0)
func compareHighCardsSlice(a, b []*Card) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	aWheel := isWheelHand(a)
	bWheel := isWheelHand(b)
	aVals := make([]int, len(a))
	bVals := make([]int, len(b))
	for i, c := range a {
		v := c.GetValue()
		if v == 1 && !aWheel {
			v = 14
		}
		aVals[i] = v
	}
	for i, c := range b {
		v := c.GetValue()
		if v == 1 && !bWheel {
			v = 14
		}
		bVals[i] = v
	}
	aTB := tieBreakValues(aVals)
	bTB := tieBreakValues(bVals)
	for i := 0; i < len(aTB) && i < len(bTB); i++ {
		if aTB[i] > bTB[i] {
			return 1
		}
		if aTB[i] < bTB[i] {
			return -1
		}
	}
	return 0
}
