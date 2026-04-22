package domain

// EnumerateTakes は hand[handIdx] を出した時に合法となる取り手を列挙する。
// 返値: それぞれ (選んだ場札インデックス, 選んだビルドインデックス) のペア。
// 数札: 場札合計 = 手札値、あるいはビルド値と一致する手札値で一致ビルドを捕獲。
// 絵札: 同ランクの場札または同ランクのビルド (絵札ビルドは作れないが念のため)。
// 1 手で場札一部 + ビルド 1 つの合成捕獲も可能。
func EnumerateTakes(handCard *Card, tableCards []*Card, builds []*CassinoBuild) [][2][]int {
	if handCard == nil {
		return nil
	}
	results := make([][2][]int, 0)

	// ビルド捕獲: ビルドの値が手札の値と一致しているものを単体 / 組合せ。
	// 絵札は合計値計算に絡まないが、ビルド値に face rank を置くケースは仕様外とする。
	matchingBuildIdxs := make([]int, 0)
	for i, b := range builds {
		if b == nil {
			continue
		}
		if CassinoIsFaceCard(handCard) {
			// 絵札でビルドを捕獲することは Cassino 標準では発生しないが、念のためスキップ。
			continue
		}
		if b.Value == CassinoCardValue(handCard) {
			matchingBuildIdxs = append(matchingBuildIdxs, i)
		}
	}

	// 場札捕獲候補 (手札ごとの合計 = target のサブセット or face 一致)
	var tableGroupsPerCard [][]int // 各グループは tableCards のインデックス
	if CassinoIsFaceCard(handCard) {
		faceIdxs := findFaceRankMatches(tableCards, handCard.GetValue())
		// 各 1 枚ずつ別グループ
		for _, idx := range faceIdxs {
			tableGroupsPerCard = append(tableGroupsPerCard, []int{idx})
		}
	} else {
		tableGroupsPerCard = findSubsetsSummingTo(tableCards, CassinoCardValue(handCard))
	}

	// 場札単体捕獲 (ビルドなし) のパターン: 1 つのグループを選ぶ / 2 つ以上の互いに disjoint なグループ
	// 実際のゲームでは同ターンで複数グループを一括捕獲可能。ここでは組合せ爆発を避けるために、
	// 「単一グループ」と「全ての disjoint グループの集合」の 2 パターンを生成する。
	if len(tableGroupsPerCard) > 0 {
		// 単一グループのみを挙げる (CPU・ヒント判定に十分)
		for _, g := range tableGroupsPerCard {
			gr := make([]int, len(g))
			copy(gr, g)
			results = append(results, [2][]int{gr, nil})
		}
		// 全ての disjoint グループをまとめて一括捕獲
		combined := combineDisjoint(tableGroupsPerCard)
		if len(combined) > 0 && len(combined) != len(tableGroupsPerCard[0]) {
			// combined と単一グループが重複しないよう、複数グループあるときに限り加える
			if len(countUniqueIdxs(combined)) > len(tableGroupsPerCard[0]) {
				results = append(results, [2][]int{combined, nil})
			}
		}
	}

	// ビルド捕獲 (場札なし)
	if len(matchingBuildIdxs) > 0 {
		bi := make([]int, len(matchingBuildIdxs))
		copy(bi, matchingBuildIdxs)
		results = append(results, [2][]int{nil, bi})
		// 場札捕獲 + ビルド捕獲の合成
		for _, g := range tableGroupsPerCard {
			gr := make([]int, len(g))
			copy(gr, g)
			biCopy := make([]int, len(matchingBuildIdxs))
			copy(biCopy, matchingBuildIdxs)
			results = append(results, [2][]int{gr, biCopy})
		}
	}
	return results
}

// combineDisjoint は disjoint なグループを全て取り込み、結合した 1 つのインデックス列にする。
func combineDisjoint(groups [][]int) []int {
	used := make(map[int]bool)
	out := make([]int, 0)
	for _, g := range groups {
		conflict := false
		for _, idx := range g {
			if used[idx] {
				conflict = true
				break
			}
		}
		if conflict {
			continue
		}
		for _, idx := range g {
			used[idx] = true
			out = append(out, idx)
		}
	}
	return out
}

// countUniqueIdxs は重複を除いたインデックス数を返す。
func countUniqueIdxs(s []int) map[int]struct{} {
	m := make(map[int]struct{}, len(s))
	for _, v := range s {
		m[v] = struct{}{}
	}
	return m
}

// EnumerateBuilds は hand[handIdx] を使って作成できるビルドの候補を列挙する。
// 宣言値は 2-10、絵札不可、かつ同値の手札が別途残っている必要がある。
func EnumerateBuilds(handCard *Card, handIdx int, hand []*Card, tableCards []*Card) []cassinoBuildCandidate {
	if handCard == nil || CassinoIsFaceCard(handCard) {
		return nil
	}
	out := make([]cassinoBuildCandidate, 0)
	for declared := 2; declared <= 10; declared++ {
		if declared <= CassinoCardValue(handCard) {
			// 宣言値は手札値より厳密に大きい (= 場札を組合せる意味がある)
			continue
		}
		// 手札に declared 値の別カードがあるか
		hasCapture := false
		for i, c := range hand {
			if i == handIdx {
				continue
			}
			if c == nil || CassinoIsFaceCard(c) {
				continue
			}
			if CassinoCardValue(c) == declared {
				hasCapture = true
				break
			}
		}
		if !hasCapture {
			continue
		}
		// 場札で sum = declared - handCardValue となる組合せを列挙
		target := declared - CassinoCardValue(handCard)
		if target <= 0 {
			continue
		}
		subsets := findSubsetsSummingTo(tableCards, target)
		for _, s := range subsets {
			cand := cassinoBuildCandidate{
				DeclaredValue: declared,
				TableIdxs:     append([]int(nil), s...),
			}
			out = append(out, cand)
		}
	}
	return out
}

// cassinoBuildCandidate はビルド列挙の結果。
type cassinoBuildCandidate struct {
	DeclaredValue int
	TableIdxs     []int
}
