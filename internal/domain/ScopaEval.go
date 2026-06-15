//go:build !js || !wasm || classic

package domain

// ScopaMaxSubsetCards は scopaSubsetsSummingTo で処理する最大カード枚数。
// スコパの場に並ぶカードは現実的には十数枚以下なので、2^n が爆発しない
// 範囲として安全側で 20 枚に抑える。
const ScopaMaxSubsetCards = 20

// ScopaCardValue スコパ用カード値を返す。
// A=1, 2-7=額面, J=8, Q=9, K=10 (カシノと異なり絵札も数値を持つ)。
func ScopaCardValue(c *Card) int {
	if c == nil {
		return 0
	}
	v := c.GetValue()
	if v <= 7 {
		return v // A(1)〜7
	}
	// J(11)=8, Q(12)=9, K(13)=10
	return v - 3
}

// ScopaIsDiamond は ♦ (デナリ) スートかどうか。
func ScopaIsDiamond(c *Card) bool {
	return c != nil && c.GetDesign() == CardDesignDiamond
}

// ScopaIsSetteBello は 7♦ (セッテ・ベッロ) かどうか。
func ScopaIsSetteBello(c *Card) bool {
	return c != nil && c.GetDesign() == CardDesignDiamond && c.GetValue() == 7
}

// ScopaIsSeven は 7 (プリミエラ簡易判定用) かどうか。
func ScopaIsSeven(c *Card) bool {
	return c != nil && c.GetValue() == 7
}

// hasSingleValueMatch は table の中に value とちょうど一致する単独カードが
// 1 枚以上存在するかを返す。スコパでは単独一致がある場合、合計値による
// 組合せ捕獲は禁止され、単独カードを取らねばならない。
func hasSingleValueMatch(tableCards []*Card, value int) bool {
	for _, c := range tableCards {
		if c != nil && ScopaCardValue(c) == value {
			return true
		}
	}
	return false
}

// scopaSubsetsSummingTo は cards の中から ScopaCardValue 合計が target になる
// 部分集合を列挙する。絵札も数値を持つため除外しない。
// 戻り値は cards に対する 0-indexed インデックスのスライスのスライス。
func scopaSubsetsSummingTo(cards []*Card, target int) [][]int {
	if target <= 0 {
		return nil
	}
	n := len(cards)
	if n == 0 || n > ScopaMaxSubsetCards {
		return nil
	}
	result := make([][]int, 0)
	for mask := 1; mask < (1 << n); mask++ {
		sum := 0
		over := false
		for i := 0; i < n; i++ {
			if mask&(1<<i) == 0 {
				continue
			}
			sum += ScopaCardValue(cards[i])
			if sum > target {
				over = true
				break
			}
		}
		if over || sum != target {
			continue
		}
		idxs := make([]int, 0)
		for i := 0; i < n; i++ {
			if mask&(1<<i) != 0 {
				idxs = append(idxs, i)
			}
		}
		result = append(result, idxs)
	}
	return result
}

// EnumerateScopaCaptures は playedCard を出したときに合法となる捕獲対象を列挙する。
// 各要素は tableCards に対するインデックス集合。
//   - 単独一致がある場合: 一致する各単独カードを 1 つずつのグループとして返す
//     (組合せ捕獲は禁止)。
//   - 単独一致がない場合: 合計値が一致する全ての部分集合を返す。
func EnumerateScopaCaptures(playedCard *Card, tableCards []*Card) [][]int {
	if playedCard == nil {
		return nil
	}
	target := ScopaCardValue(playedCard)
	if hasSingleValueMatch(tableCards, target) {
		out := make([][]int, 0)
		for i, c := range tableCards {
			if c != nil && ScopaCardValue(c) == target {
				out = append(out, []int{i})
			}
		}
		return out
	}
	return scopaSubsetsSummingTo(tableCards, target)
}

// isValidScopaCapture は選ばれた場札インデックスが playedCard で合法に捕獲できるかを検証する。
//   - インデックスは範囲内かつ重複なし。
//   - 単独一致がある場合: 選択はちょうど 1 枚で、その値が playedCard と一致。
//   - 単独一致がない場合: 選択カードの合計値が playedCard の値に一致 (1 枚以上)。
func isValidScopaCapture(playedCard *Card, tableCards []*Card, selectedIdxs []int) bool {
	if playedCard == nil || len(selectedIdxs) == 0 {
		return false
	}
	seen := make(map[int]bool, len(selectedIdxs))
	chosen := make([]*Card, 0, len(selectedIdxs))
	for _, idx := range selectedIdxs {
		if idx < 0 || idx >= len(tableCards) || seen[idx] {
			return false
		}
		seen[idx] = true
		chosen = append(chosen, tableCards[idx])
	}

	target := ScopaCardValue(playedCard)
	if hasSingleValueMatch(tableCards, target) {
		// 単独一致が存在する場合は単独カードのみ捕獲できる。
		return len(chosen) == 1 && ScopaCardValue(chosen[0]) == target
	}

	sum := 0
	for _, c := range chosen {
		sum += ScopaCardValue(c)
	}
	return sum == target
}
