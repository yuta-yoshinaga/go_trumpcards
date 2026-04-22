package domain

// CassinoMaxSubsetCards は findSubsetsSummingTo で処理する最大カード枚数。
// 場に並ぶ可能性のある枚数 (48 ≒ デッキ全部) を大きく上回ると 2^n が現実的でなくなるため、
// 安全側で 20 枚以下に抑える。カシノの実戦で場に 20 枚以上同時に並ぶことはない。
const CassinoMaxSubsetCards = 20

// CassinoCardValue カシノ用カード値 (A=1, 2-10=face, J=11, Q=12, K=13)
func CassinoCardValue(c *Card) int {
	if c == nil {
		return 0
	}
	v := c.GetValue()
	if v == 1 {
		return 1 // Ace は 1 固定
	}
	return v
}

// CassinoIsFaceCard は絵札 (J/Q/K) かどうか。絵札は合計値計算に使わずランク一致のみ。
func CassinoIsFaceCard(c *Card) bool {
	if c == nil {
		return false
	}
	v := c.GetValue()
	return v >= 11 && v <= 13
}

// CassinoIsSpade は ♠ スートかどうか。
func CassinoIsSpade(c *Card) bool {
	return c != nil && c.GetDesign() == CardDesignSpade
}

// CassinoIsBigCasino は 10♦ (ビッグカシノ) かどうか。
func CassinoIsBigCasino(c *Card) bool {
	return c != nil && c.GetDesign() == CardDesignDiamond && c.GetValue() == 10
}

// CassinoIsLittleCasino は 2♠ (リトルカシノ) かどうか。
func CassinoIsLittleCasino(c *Card) bool {
	return c != nil && c.GetDesign() == CardDesignSpade && c.GetValue() == 2
}

// CassinoIsAce は A (エース) かどうか。
func CassinoIsAce(c *Card) bool {
	return c != nil && c.GetValue() == 1
}

// findSubsetsSummingTo は cards の中から合計値が target になる部分集合を列挙する。
// - 絵札 (J/Q/K) を含む部分集合は除外 (絵札はランク一致のみで捕獲できる)。
// - target <= 0 の場合や cards が多すぎる場合は nil を返す。
// 戻り値は cards に対する 0-indexed インデックスのスライスのスライス。
// cribbage_scoring.go::CribbageScoreFifteens のビットマスク列挙パターンを流用。
func findSubsetsSummingTo(cards []*Card, target int) [][]int {
	if target <= 0 {
		return nil
	}
	n := len(cards)
	if n == 0 || n > CassinoMaxSubsetCards {
		return nil
	}
	result := make([][]int, 0)
	for mask := 1; mask < (1 << n); mask++ {
		sum := 0
		skip := false
		for i := 0; i < n; i++ {
			if mask&(1<<i) == 0 {
				continue
			}
			if CassinoIsFaceCard(cards[i]) {
				skip = true
				break
			}
			sum += CassinoCardValue(cards[i])
			if sum > target {
				skip = true
				break
			}
		}
		if skip || sum != target {
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

// findFaceRankMatches は cards の中から rank (11=J, 12=Q, 13=K) と一致する
// カードのインデックスを返す。絵札のランク一致捕獲で使う。
func findFaceRankMatches(cards []*Card, rank int) []int {
	idxs := make([]int, 0)
	for i, c := range cards {
		if c == nil {
			continue
		}
		if c.GetValue() == rank {
			idxs = append(idxs, i)
		}
	}
	return idxs
}

// isValidTakeSelection は選ばれた場札のインデックスが、playedCard の値に対して
// 1 つ以上の合計一致グループでちょうど分割できるかを検証する。
// 絵札の場合はランク一致のみで、指定された場札全てが同ランクであること。
// 成功時 true と、実際に使われた (グループごとの) インデックス配列を返す。
func isValidTakeSelection(playedCard *Card, tableCards []*Card, selectedIdxs []int) (bool, [][]int) {
	if playedCard == nil || len(selectedIdxs) == 0 {
		return false, nil
	}
	// インデックス重複・範囲チェック
	seen := make(map[int]bool, len(selectedIdxs))
	chosen := make([]*Card, 0, len(selectedIdxs))
	for _, idx := range selectedIdxs {
		if idx < 0 || idx >= len(tableCards) || seen[idx] {
			return false, nil
		}
		seen[idx] = true
		chosen = append(chosen, tableCards[idx])
	}

	// 絵札の場合: 選ばれた全てが絵札かつ playedCard と同ランク
	if CassinoIsFaceCard(playedCard) {
		for _, c := range chosen {
			if !CassinoIsFaceCard(c) || c.GetValue() != playedCard.GetValue() {
				return false, nil
			}
		}
		// 絵札捕獲はカード単独を 1 グループとして扱う
		groups := make([][]int, 0, len(selectedIdxs))
		for _, idx := range selectedIdxs {
			groups = append(groups, []int{idx})
		}
		return true, groups
	}

	// 数札の場合: 合計値が playedCard の値に一致するグループで分割できるか
	target := CassinoCardValue(playedCard)
	ok, groups := partitionIntoSumGroups(chosen, selectedIdxs, target)
	return ok, groups
}

// partitionIntoSumGroups は chosen カード群を target 値のサブセットに分割できるか試す。
// 全カードを使い切って target 合計のグループを 1 つ以上作ることが条件。
// 絵札を含む入力では false。
func partitionIntoSumGroups(chosen []*Card, origIdxs []int, target int) (bool, [][]int) {
	if len(chosen) == 0 || target <= 0 {
		return false, nil
	}
	for _, c := range chosen {
		if CassinoIsFaceCard(c) {
			return false, nil
		}
	}
	// chosen 内の絶対位置 (0..len(chosen)-1) を origIdxs に変換
	used := make([]bool, len(chosen))
	groups := make([][]int, 0)
	if !dfsPartition(chosen, used, target, &groups, origIdxs) {
		return false, nil
	}
	if len(groups) == 0 {
		return false, nil
	}
	return true, groups
}

// dfsPartition は chosen を target 合計のグループに分割する DFS。
// 使い切れば true。
func dfsPartition(chosen []*Card, used []bool, target int, groups *[][]int, origIdxs []int) bool {
	// 最初の未使用インデックスを探す
	start := -1
	for i, u := range used {
		if !u {
			start = i
			break
		}
	}
	if start == -1 {
		return true // 全部使い切った
	}
	// start を含むグループを探す (target 合計になる部分集合)
	remainIdxs := make([]int, 0)
	remainCards := make([]*Card, 0)
	remainIdxs = append(remainIdxs, start)
	remainCards = append(remainCards, chosen[start])
	for i := start + 1; i < len(chosen); i++ {
		if !used[i] {
			remainIdxs = append(remainIdxs, i)
			remainCards = append(remainCards, chosen[i])
		}
	}
	// start を含むサブセットのみ列挙 (bit 0 を固定 1)
	n := len(remainCards)
	if n > CassinoMaxSubsetCards {
		return false
	}
	for mask := 1; mask < (1 << n); mask++ {
		if mask&1 == 0 {
			continue
		}
		sum := 0
		for i := 0; i < n; i++ {
			if mask&(1<<i) != 0 {
				sum += CassinoCardValue(remainCards[i])
			}
		}
		if sum != target {
			continue
		}
		// このマスクを使って再帰
		picked := make([]int, 0)
		for i := 0; i < n; i++ {
			if mask&(1<<i) != 0 {
				used[remainIdxs[i]] = true
				picked = append(picked, origIdxs[remainIdxs[i]])
			}
		}
		*groups = append(*groups, picked)
		if dfsPartition(chosen, used, target, groups, origIdxs) {
			return true
		}
		// 戻す
		for i := 0; i < n; i++ {
			if mask&(1<<i) != 0 {
				used[remainIdxs[i]] = false
			}
		}
		*groups = (*groups)[:len(*groups)-1]
	}
	return false
}
