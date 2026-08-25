//go:build !js || !wasm || extra

package domain

// コストリー・カラーズの得点評価。
//
// **#5461 はクリブが無いことしか書いていない。** 実際に Cribbage と違うのは
// もっと広い:
//
//   - 配るのは **3 枚だけ**で、次の 1 枚を「トランプ」として表に返す
//   - 数え上げは 15 と 31 のほかに **25** でも点になる ── この 25 が
//     Cribbage には無い
//   - プライアル (同位 3 枚) は **9**、ダブルプライアル (4 枚) は **18**。
//     Cribbage の 6 / 12 ではない
//   - ショーで数えるのは「フィフティーン・ペア・ラン・フラッシュ」ではなく、
//     **色とスートの梯子** ── これがゲーム名の由来
//   - J と 2 は特別で、トランプと同じスートなら 4 点、それ以外でも 2 点
//
// **ショーは手札 3 枚 + 表向きの 1 枚 = 4 枚で数える。** 「4 枚同スート」の
// 役が手札 3 枚だけでは成立しないので、Cribbage のスターターと同じく表の
// 1 枚が加わると読むほかない。

// 数え上げの節目。**25 が Cribbage との分かれ目。**
const (
	// CostlyFifteen は 15 の節目。
	CostlyFifteen = 15
	// CostlyTwentyFive は 25 の節目。**Cribbage には無い。**
	CostlyTwentyFive = 25
	// CostlyThirtyOne は 31 の節目 (上限でもある)。
	CostlyThirtyOne = 31
)

// 得点。
const (
	// CostlyPairPoints はペア (同位 2 枚) の点。
	CostlyPairPoints = 2
	// CostlyPrialPoints はプライアル (同位 3 枚) の点。**Cribbage の 6 ではない。**
	CostlyPrialPoints = 9
	// CostlyDoublePrialPoints はダブルプライアル (同位 4 枚) の点。
	CostlyDoublePrialPoints = 18
	// CostlyHeelsPoints は J が表に返ったときの親の点 ("for his heels")。
	CostlyHeelsPoints = 4
	// CostlyTrumpJackDeucePoints は手札のトランプスートの J / 2 の点。
	CostlyTrumpJackDeucePoints = 4
	// CostlyPlainJackDeucePoints は手札のそれ以外の J / 2 の点 ("for his nob")。
	CostlyPlainJackDeucePoints = 2
	// CostlyGoPoints は相手が「ゴー」を宣言したときの点。
	CostlyGoPoints = 1
	// CostlyLatterPoints は 31 に届かず最後の札を出したときの点 ("the latter")。
	CostlyLatterPoints = 1
	// CostlyMogRefusalPoints は交換を断られた側に入る点。
	CostlyMogRefusalPoints = 1
)

// 色とスートの梯子。**ゲーム名はこの一番上から来ている。**
const (
	// CostlyThreeInColourPoints は 3 枚同色 (スートは 2 種) の点。
	CostlyThreeInColourPoints = 2
	// CostlyThreeInSuitPoints は 3 枚同スートの点。
	CostlyThreeInSuitPoints = 3
	// CostlyFourInColourTwoSuitPoints は 4 枚同色で 2 枚が同スートの点。
	CostlyFourInColourTwoSuitPoints = 4
	// CostlyFourInColourThreeSuitPoints は 4 枚同色で 3 枚が同スートの点。
	CostlyFourInColourThreeSuitPoints = 5
	// CostlyColoursPoints は 4 枚同スート ──「コストリー・カラーズ」の点。
	CostlyColoursPoints = 6
)

// 色とスートの役の識別子。
const (
	CostlyComboNone              = ""
	CostlyComboThreeInColour     = "threeInColour"
	CostlyComboThreeInSuit       = "threeInSuit"
	CostlyComboFourInColourTwo   = "fourInColourTwoSuit"
	CostlyComboFourInColourThree = "fourInColourThreeSuit"
	CostlyComboCostlyColours     = "costlyColours"
)

// CostlyIsRed は赤いスートかを返す。
func CostlyIsRed(c *Card) bool {
	return c != nil && (c.GetDesign() == CardDesignHeart || c.GetDesign() == CardDesignDiamond)
}

// CostlyIsJackOrDeuce は J か 2 かを返す。**この 2 つだけが特別扱い。**
func CostlyIsJackOrDeuce(c *Card) bool {
	return c != nil && (c.GetValue() == 11 || c.GetValue() == 2)
}

// CostlyCardValue は数え上げに使う値を返す。
//
// **A は 1、絵札は 10。** Cribbage と同じで、A が高いのは順位の話であって
// 数え上げの値ではない。
func CostlyCardValue(c *Card) int {
	if c == nil {
		return 0
	}
	v := c.GetValue()
	if v > 10 {
		return 10
	}
	return v
}

// CostlyColourCombo は 4 枚 (手札 3 + 表の 1) の色とスートの役を返す。
//
// **梯子は上から順に見る。** 4 枚同スートが一番上で、下ほど緩い。
func CostlyColourCombo(cards []*Card) (string, int) {
	if len(cards) != 4 {
		return CostlyComboNone, 0
	}
	suits := map[int]int{}
	reds := 0
	for _, c := range cards {
		if c == nil {
			return CostlyComboNone, 0
		}
		suits[c.GetDesign()]++
		if CostlyIsRed(c) {
			reds++
		}
	}
	allOneColour := reds == 0 || reds == 4
	maxSuit := 0
	for _, n := range suits {
		if n > maxSuit {
			maxSuit = n
		}
	}
	switch {
	case maxSuit == 4:
		return CostlyComboCostlyColours, CostlyColoursPoints
	case allOneColour && maxSuit == 3:
		return CostlyComboFourInColourThree, CostlyFourInColourThreeSuitPoints
	case allOneColour:
		return CostlyComboFourInColourTwo, CostlyFourInColourTwoSuitPoints
	case maxSuit == 3:
		return CostlyComboThreeInSuit, CostlyThreeInSuitPoints
	}
	// 3 枚同色 (かつ 3 枚同スートではない)。
	if reds == 3 || reds == 1 {
		return CostlyComboThreeInColour, CostlyThreeInColourPoints
	}
	return CostlyComboNone, 0
}

// CostlyRankCombo は 4 枚の同位役 (ペア / プライアル / ダブルプライアル) を返す。
//
// **J と 2 は数えない。** その 2 つは別枠で点になるので、ここで二重に数えない。
func CostlyRankCombo(cards []*Card) (int, int) {
	counts := map[int]int{}
	for _, c := range cards {
		if c == nil || CostlyIsJackOrDeuce(c) {
			continue
		}
		counts[c.GetValue()]++
	}
	best, bestPts := 0, 0
	for _, n := range counts {
		pts := 0
		switch n {
		case 2:
			pts = CostlyPairPoints
		case 3:
			pts = CostlyPrialPoints
		case 4:
			pts = CostlyDoublePrialPoints
		}
		if pts > bestPts {
			best, bestPts = n, pts
		}
	}
	return best, bestPts
}

// CostlyJackDeucePoints は手札の J / 2 の点を返す。
//
// **トランプと同じスートなら 4、それ以外は 2。** 2 枚あれば 6、片方が
// トランプなら 8 になる ── 個別に足した結果と一致するので、そのまま足す。
func CostlyJackDeucePoints(hand []*Card, trumpDesign int) int {
	total := 0
	for _, c := range hand {
		if !CostlyIsJackOrDeuce(c) {
			continue
		}
		if c.GetDesign() == trumpDesign {
			total += CostlyTrumpJackDeucePoints
			continue
		}
		total += CostlyPlainJackDeucePoints
	}
	return total
}

// CostlyRunLength は連なりの末尾から数えた最長の階段の長さを返す (3 未満は 0)。
//
// **並び順は問わない。** 5-7-6 と出ても 3 枚の階段として数える。
func CostlyRunLength(pile []*Card) int {
	for n := len(pile); n >= 3; n-- {
		if costlyIsRun(pile[len(pile)-n:]) {
			return n
		}
	}
	return 0
}

// costlyIsRun は与えられた札が (順不同で) 連番かを返す。
func costlyIsRun(cards []*Card) bool {
	seen := map[int]bool{}
	lo, hi := 1<<31-1, 0
	for _, c := range cards {
		if c == nil {
			return false
		}
		v := c.GetValue()
		if seen[v] {
			return false
		}
		seen[v] = true
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}
	return hi-lo == len(cards)-1
}

// CostlyPlayScore は 1 枚出した直後の得点と、その理由の識別子を返す。
func CostlyPlayScore(pile []*Card, total int) (int, []string) {
	pts := 0
	reasons := make([]string, 0, 3)

	// **15・25・31 はどれも「作った枚数」ぶん点になる。**
	switch total {
	case CostlyFifteen:
		pts += len(pile)
		reasons = append(reasons, "fifteen")
	case CostlyTwentyFive:
		pts += len(pile)
		reasons = append(reasons, "twentyFive")
	case CostlyThirtyOne:
		pts += len(pile)
		reasons = append(reasons, "thirtyOne")
	}

	if n, p := costlyTailRankCombo(pile); p > 0 {
		pts += p
		reasons = append(reasons, costlyRankReason(n))
	}
	if n := CostlyRunLength(pile); n > 0 {
		pts += n
		reasons = append(reasons, "run")
	}
	return pts, reasons
}

// costlyTailRankCombo は連なりの末尾に並んだ同位の枚数と点を返す。
func costlyTailRankCombo(pile []*Card) (int, int) {
	if len(pile) == 0 {
		return 0, 0
	}
	v := pile[len(pile)-1].GetValue()
	n := 0
	for i := len(pile) - 1; i >= 0 && pile[i].GetValue() == v; i-- {
		n++
	}
	switch n {
	case 2:
		return n, CostlyPairPoints
	case 3:
		return n, CostlyPrialPoints
	case 4:
		return n, CostlyDoublePrialPoints
	}
	return 0, 0
}

// costlyRankReason は同位役の理由識別子を返す。
func costlyRankReason(n int) string {
	switch n {
	case 3:
		return "prial"
	case 4:
		return "doublePrial"
	default:
		return "pair"
	}
}
