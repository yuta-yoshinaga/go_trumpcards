package domain

// CassinoHintAction はヒントが推奨する行動種別。
type CassinoHintAction int

const (
	// CassinoHintTrail 捕獲できないので場に置く。
	CassinoHintTrail CassinoHintAction = iota
	// CassinoHintTake 場札 / ビルドを捕獲する。
	CassinoHintTake
	// CassinoHintBuild ビルドを作る。
	CassinoHintBuild
)

// CassinoHint は CUI 向けの推奨手。
type CassinoHint struct {
	Action    CassinoHintAction
	HandIdx   int   // 出す手札のインデックス
	TableIdxs []int // take: 捕獲する場札 (loose card) のインデックス
	BuildIdxs []int // take: 捕獲するビルドのインデックス
	Value     int   // take: 捕獲値 / build: 宣言値
}

// SuggestCassinoMove は手札・場札・ビルドから推奨手を返す。
// 優先度は「捕獲枚数が最大の take > build > trail」。take で示す場札の部分集合は
// 必ず手札の値ちょうどに合計するため、常に合法な捕獲となる (最適とは限らない)。
// hand が空なら nil を返す。
func SuggestCassinoMove(hand, table []*Card, builds []*CassinoBuild) *CassinoHint {
	if len(hand) == 0 {
		return nil
	}

	var best *CassinoHint
	bestCaptured := 0
	for hi, hc := range hand {
		if hc == nil {
			continue
		}
		tableIdxs, buildIdxs, captured := cassinoBestCaptureFor(hc, table, builds)
		if captured > 0 && captured > bestCaptured {
			bestCaptured = captured
			best = &CassinoHint{
				Action:    CassinoHintTake,
				HandIdx:   hi,
				TableIdxs: tableIdxs,
				BuildIdxs: buildIdxs,
				Value:     CassinoCardValue(hc),
			}
		}
	}
	if best != nil {
		return best
	}

	if b := cassinoBestBuild(hand, table); b != nil {
		return b
	}

	return &CassinoHint{Action: CassinoHintTrail, HandIdx: cassinoLowestHandIdx(hand)}
}

// cassinoBestCaptureFor は 1 枚の手札で捕獲できる場札 / ビルドと捕獲枚数を返す。
func cassinoBestCaptureFor(hc *Card, table []*Card, builds []*CassinoBuild) (tableIdxs, buildIdxs []int, captured int) {
	if CassinoIsFaceCard(hc) {
		// 絵札は同ランクの場札のみ捕獲。ビルドは数値なので対象外。
		idxs := findFaceRankMatches(table, hc.GetValue())
		return idxs, nil, len(idxs)
	}

	v := CassinoCardValue(hc)
	// 最も多く捕獲できる loose card の部分集合を選ぶ。
	var bestSub []int
	for _, s := range findSubsetsSummingTo(table, v) {
		if len(s) > len(bestSub) {
			bestSub = s
		}
	}
	// 宣言値が一致するビルドも同時に捕獲できる。
	var bidxs []int
	buildCap := 0
	for bi, b := range builds {
		if b != nil && b.Value == v {
			bidxs = append(bidxs, bi)
			buildCap += len(b.AllCards())
		}
	}
	return bestSub, bidxs, len(bestSub) + buildCap
}

// cassinoBestBuild は手札の数札を loose card と組み合わせてビルドを作れるかを探す。
// 合成値 target を宣言するには target と同値の手札をもう 1 枚持っている必要がある。
func cassinoBestBuild(hand, table []*Card) *CassinoHint {
	for hi, hc := range hand {
		if hc == nil || CassinoIsFaceCard(hc) {
			continue
		}
		v := CassinoCardValue(hc)
		for target := v + 1; target <= 10; target++ {
			subs := findSubsetsSummingTo(table, target-v)
			if len(subs) == 0 {
				continue
			}
			if cassinoHandHasValueExcept(hand, target, hi) {
				return &CassinoHint{Action: CassinoHintBuild, HandIdx: hi, TableIdxs: subs[0], Value: target}
			}
		}
	}
	return nil
}

// cassinoHandHasValueExcept は except 以外の手札に値 v の数札があるかを返す。
func cassinoHandHasValueExcept(hand []*Card, v, except int) bool {
	for i, c := range hand {
		if i == except || c == nil || CassinoIsFaceCard(c) {
			continue
		}
		if CassinoCardValue(c) == v {
			return true
		}
	}
	return false
}

// cassinoLowestHandIdx は最も値の小さい手札のインデックスを返す (trail 用)。
func cassinoLowestHandIdx(hand []*Card) int {
	best := -1
	bestVal := 1 << 30
	for i, c := range hand {
		if c == nil {
			continue
		}
		v := c.GetValue()
		if v < bestVal {
			bestVal = v
			best = i
		}
	}
	if best < 0 {
		return 0
	}
	return best
}
