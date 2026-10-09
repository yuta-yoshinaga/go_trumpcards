//go:build !js || !wasm || extra8

package domain

// ビデオポーカーの推奨ホールドを、種目ごとの戦略表 (優先順リスト) で決める (#10969)。
//
// 実行時に期待値は計算しない。ワイルド入りの全ドロー列挙は Worker の CPU 予算
// (10ms) に収まらないので、最大ベット (ナチュラルロイヤル 800 倍) 前提の戦略表を
// 持ち、表の順序はオフラインの検証器 (VideoPokerEV_vpev_test.go、`-tags "test vpev"`)
// で正確な期待値と突き合わせて決めている。
//
// 32 通りのホールド候補を表の行ごとに調べ、条件を満たす候補がある最初の行を採る。
// ワイルド種目ではワイルドを常に残し、表はワイルドの枚数ごとに持つ。

// VideoPokerHoldAdvice は推奨ホールドと、その根拠になった戦略表の行。
// RuleKey は i18n キー videopoker.strategy.<RuleKey> の末尾。
type VideoPokerHoldAdvice struct {
	Hold    [VideoPokerHandSize]bool
	RuleKey string
}

// vpRuleKind は戦略表の 1 行が見るホールドの形。
type vpRuleKind int

const (
	vpMade          vpRuleKind = iota // 5 枚全部が cats のどれか
	vpFourDeuces                      // 2 を 4 枚 (Deuces Wild)
	vpQuads                           // 4 枚組 (ワイルド込み)
	vpTrips                           // 3 枚組 (ワイルド込み)
	vpTwoPair                         // 自然札の 2 ペア
	vpHighPair                        // minRank 以上の 1 ペア (ワイルド 1 枚 + 高位札 1 枚も含む)
	vpLowPair                         // minRank 未満の自然札の 1 ペア
	vpOnePair                         // 自然札の 1 ペア (2 ペアなら高い方)
	vpRoyalDraw                       // n 枚 (ワイルド込み) がロイヤルに向かう
	vpSFDraw                          // n 枚 (ワイルド込み) がストレートフラッシュに向かう
	vpFlushDraw                       // n 枚 (ワイルド込み) が同スート
	vpOutside                         // 4 枚 (ワイルド込み) で完成ランクが 2 つ以上のストレート待ち
	vpInside                          // 4 枚 (ワイルド込み) で完成ランクが 1 つのストレート待ち
	vpSuitedHigh                      // minRank 以上の同スート 2 枚
	vpSuitedTenHigh                   // 10 と J/Q/K の同スート 2 枚
	vpUnsuitedHigh                    // minRank 以上の札 2 枚 (3 枚以上あれば低い 2 枚)
	vpOneHigh                         // minRank 以上の札 1 枚
	vpWildsOnly                       // ワイルドだけ
	vpWildPlusOne                     // ワイルド + minRank..maxRank の自然札 1 枚
)

// 役の区分 (vpMade の cats)。
const (
	vpCatNaturalRoyal = iota
	vpCatWildRoyal
	vpCatFiveOfAKind
	vpCatStraightFlush
	vpCatFullHouse
	vpCatFlush
	vpCatStraight
)

type vpRule struct {
	key  string
	kind vpRuleKind
	// n はホールドの枚数 (ワイルド込み)。0 は枚数を問わない。
	n int
	// minRank は高位札の下限 (A=14)。SF ドローでは自然札の最低ランクの下限。
	minRank int
	// consecutive は SF ドローで自然札が連続ランクであること。
	consecutive bool
	// maxRank はロイヤルドローの自然札の上限 (0 は上限なし、A=14)。
	// Deuces Wild の 2 枚ロイヤルは 10/J/Q だけが全部捨てるより伸びる。
	maxRank int
	// noAceLow は A を下 (A-5 の窓) で使うストレート系を除く。
	noAceLow bool
	// maxSpan は SF ドローの自然札の最高と最低の差の上限 (0 は上限なし)。
	maxSpan int
	// needHigh はロイヤルドローに needHigh 以上の自然札が 1 枚以上要ること (0 は条件なし)。
	needHigh int
	cats     []int
}

func vpMadeRule(cats ...int) vpRule { return vpRule{key: "madeHand", kind: vpMade, cats: cats} }

var (
	vpRuleQuads       = vpRule{key: "keepQuads", kind: vpQuads, n: 4}
	vpRuleTrips       = vpRule{key: "keepTrips", kind: vpTrips, n: 3}
	vpRuleTwoPair     = vpRule{key: "keepTwoPair", kind: vpTwoPair, n: 4}
	vpRuleRoyal4      = vpRule{key: "royalDraw4", kind: vpRoyalDraw, n: 4}
	vpRuleRoyal3      = vpRule{key: "royalDraw3", kind: vpRoyalDraw, n: 3}
	vpRuleRoyal2      = vpRule{key: "royalDraw2", kind: vpRoyalDraw, n: 2}
	vpRuleSF4         = vpRule{key: "straightFlushDraw4", kind: vpSFDraw, n: 4}
	vpRuleSF3         = vpRule{key: "straightFlushDraw3", kind: vpSFDraw, n: 3}
	vpRuleFlush4      = vpRule{key: "flushDraw4", kind: vpFlushDraw, n: 4}
	vpRuleOutside     = vpRule{key: "outsideStraightDraw", kind: vpOutside, n: 4}
	vpRuleSuitedTen   = vpRule{key: "suitedTenHigh", kind: vpSuitedTenHigh, n: 2}
	vpRuleWildsOnly   = vpRule{key: "keepWilds", kind: vpWildsOnly}
	vpRuleFourDeuces  = vpRule{key: "keepFourDeuces", kind: vpFourDeuces, n: 4}
	vpRuleOnePairOnly = vpRule{key: "keepPair", kind: vpOnePair, n: 2}
)

func vpHighPairRule(minRank int) vpRule {
	return vpRule{key: "keepHighPair", kind: vpHighPair, n: 2, minRank: minRank}
}

func vpLowPairRule(minRank int) vpRule {
	return vpRule{key: "keepLowPair", kind: vpLowPair, n: 2, minRank: minRank}
}

func vpSuitedHighRule(minRank int) vpRule {
	return vpRule{key: "suitedHighCards", kind: vpSuitedHigh, n: 2, minRank: minRank}
}

func vpUnsuitedHighRule(minRank int) vpRule {
	return vpRule{key: "highCards", kind: vpUnsuitedHigh, n: 2, minRank: minRank}
}

func vpOneHighRule(minRank int) vpRule {
	return vpRule{key: "oneHighCard", kind: vpOneHigh, n: 1, minRank: minRank}
}

func vpInsideRule(minHigh int) vpRule {
	return vpRule{key: "insideStraightDraw", kind: vpInside, n: 4, minRank: minHigh}
}

func vpSFRule(n, minRank int, consecutive bool) vpRule {
	r := vpRuleSF4
	if n == 3 {
		r = vpRuleSF3
	}
	r.minRank = minRank
	r.consecutive = consecutive
	return r
}

// vpStrategy は variant とワイルド枚数に対する戦略表を返す。
func vpStrategy(variant string, wilds int) []vpRule {
	switch variant {
	case "deuceswild":
		return vpDeucesWildStrategy(wilds)
	case "jokerpoker":
		return vpJokerPokerStrategy(wilds)
	default:
		return vpJacksOrBetterStrategy()
	}
}

// vpJacksOrBetterStrategy は 9/6 Jacks or Better の戦略表。
func vpJacksOrBetterStrategy() []vpRule {
	const high = 11
	return []vpRule{
		vpMadeRule(vpCatNaturalRoyal, vpCatStraightFlush),
		vpRuleQuads,
		vpRuleRoyal4,
		vpMadeRule(vpCatFullHouse, vpCatFlush, vpCatStraight),
		vpRuleTrips,
		vpRuleSF4,
		vpRuleTwoPair,
		vpHighPairRule(high),
		vpRuleRoyal3,
		vpRuleFlush4,
		vpLowPairRule(high),
		vpRuleOutside,
		vpSuitedHighRule(high),
		vpRuleSF3,
		vpInsideRule(high),
		vpUnsuitedHighRule(high),
		vpRuleSuitedTen,
		vpOneHighRule(high),
	}
}

// vpDeucesWildStrategy は full pay Deuces Wild (25/15/9/5/3/2/2/1) の戦略表。
func vpDeucesWildStrategy(wilds int) []vpRule {
	switch wilds {
	case 4:
		return []vpRule{vpRuleFourDeuces}
	case 3:
		return []vpRule{vpMadeRule(vpCatWildRoyal, vpCatFiveOfAKind), vpRuleWildsOnly}
	case 2:
		return []vpRule{
			vpMadeRule(vpCatWildRoyal, vpCatFiveOfAKind, vpCatStraightFlush),
			vpRuleQuads,
			vpRuleRoyal4,
			vpSFRule(4, 6, true),
			vpRuleWildsOnly,
		}
	case 1:
		return []vpRule{
			vpMadeRule(vpCatWildRoyal, vpCatFiveOfAKind, vpCatStraightFlush),
			vpRuleQuads,
			vpRuleRoyal4,
			vpMadeRule(vpCatFullHouse),
			vpSFRule(4, 5, true),
			vpRuleTrips,
			vpMadeRule(vpCatFlush, vpCatStraight),
			vpRuleSF4,
			vpRule{key: "royalDraw3", kind: vpRoyalDraw, n: 3, maxRank: 13},
			vpSFRule(3, 6, true),
			vpRuleWildsOnly,
		}
	default:
		return []vpRule{
			vpMadeRule(vpCatNaturalRoyal),
			vpRuleRoyal4,
			vpMadeRule(vpCatStraightFlush),
			vpRuleQuads,
			vpMadeRule(vpCatFullHouse, vpCatFlush),
			vpRuleTrips,
			vpMadeRule(vpCatStraight),
			vpRuleSF4,
			vpRuleRoyal3,
			vpRuleOnePairOnly,
			vpRuleFlush4,
			vpRuleOutside,
			vpSFRule(3, 2, false),
			vpRule{key: "royalDraw2", kind: vpRoyalDraw, n: 2, maxRank: 12},
			vpRule{key: "insideStraightDraw", kind: vpInside, n: 4, noAceLow: true},
		}
	}
}

// vpJokerPokerStrategy は 20/7/5 Joker Poker (Kings or Better) の戦略表。
func vpJokerPokerStrategy(wilds int) []vpRule {
	const high = 13
	if wilds > 0 {
		return []vpRule{
			vpMadeRule(vpCatWildRoyal, vpCatFiveOfAKind, vpCatStraightFlush),
			vpRuleQuads,
			vpRuleRoyal4,
			vpMadeRule(vpCatFullHouse),
			vpSFRule(4, 0, true),
			vpRuleTrips,
			vpMadeRule(vpCatFlush, vpCatStraight),
			vpRuleSF4,
			vpRuleRoyal3,
			vpSFRule(3, 0, true),
			vpRule{key: "flushDraw4", kind: vpFlushDraw, n: 4, needHigh: high},
			vpHighPairRule(high),
			vpRule{key: "straightFlushDraw3", kind: vpSFDraw, n: 3, maxSpan: 3},
			vpRuleFlush4,
			vpRule{key: "keepWildAndMiddleCard", kind: vpWildPlusOne, n: 2, minRank: 6, maxRank: 10},
			vpRuleWildsOnly,
		}
	}
	return []vpRule{
		vpMadeRule(vpCatNaturalRoyal, vpCatStraightFlush),
		vpRuleQuads,
		vpRuleRoyal4,
		vpMadeRule(vpCatFullHouse, vpCatFlush, vpCatStraight),
		vpRuleTrips,
		vpRuleSF4,
		vpRuleTwoPair,
		vpHighPairRule(high),
		vpRuleRoyal3,
		vpRuleFlush4,
		vpLowPairRule(high),
		vpSFRule(3, 0, true),
		vpRuleOutside,
		vpRule{key: "royalDraw2", kind: vpRoyalDraw, n: 2, needHigh: high},
		vpRuleSF3,
		vpUnsuitedHighRule(high),
		vpOneHighRule(high),
		vpRuleRoyal2,
	}
}

// VideoPokerStrategyRuleKeys は戦略表が返し得る RuleKey を全部返す (i18n の照合用)。
func VideoPokerStrategyRuleKeys() []string {
	seen := map[string]bool{"drawAll": true}
	keys := []string{"drawAll"}
	for _, variant := range []string{"jacksorbetter", "deuceswild", "jokerpoker"} {
		for wilds := 0; wilds <= 4; wilds++ {
			for _, r := range vpStrategy(variant, wilds) {
				if !seen[r.key] {
					seen[r.key] = true
					keys = append(keys, r.key)
				}
			}
		}
	}
	return keys
}

// vpCard はホールド判定用の札。rank は 2..14 (A=14)。
type vpCard struct {
	rank, suit int
	wild       bool
}

// vpHold は 1 つのホールド候補の中身。
type vpHold struct {
	mask    int
	wilds   int
	natural []vpCard
}

func (h vpHold) size() int { return h.wilds + len(h.natural) }

// RecommendVideoPokerHold は variant ("jacksorbetter" | "deuceswild" | "jokerpoker"、
// 未知は jacksorbetter) の戦略表で hand (5 枚) の推奨ホールドを返す。
// 5 枚でない手は全部捨てる (RuleKey "drawAll")。
func RecommendVideoPokerHold(variant string, hand []*Card) VideoPokerHoldAdvice {
	advice, _ := vpRecommend(variant, hand)
	return advice
}

// vpRecommend は推奨と、それを決めた戦略表の行のインデックスを返す。
// どの行にも当たらず全部捨てるとき (5 枚でない手を含む) は -1。
func vpRecommend(variant string, hand []*Card) (VideoPokerHoldAdvice, int) {
	drawAll := VideoPokerHoldAdvice{RuleKey: "drawAll"}
	if len(hand) != VideoPokerHandSize {
		return drawAll, -1
	}
	cfg := vpConfigFor(variant)
	cards := make([]vpCard, len(hand))
	wildMask, wilds := 0, 0
	for i, c := range hand {
		if c == nil {
			return drawAll, -1
		}
		w := cfg.IsWild != nil && cfg.IsWild(c)
		cards[i] = vpCard{rank: vpRank(c), suit: c.GetDesign(), wild: w}
		if w {
			wildMask |= 1 << i
			wilds++
		}
	}
	cat := vpMadeCategory(cfg, hand)
	for row, r := range vpStrategy(cfg.Name, wilds) {
		if mask, ok := vpBestMask(r, cards, wildMask, cat); ok {
			advice := VideoPokerHoldAdvice{RuleKey: r.key}
			for i := range hand {
				advice.Hold[i] = mask&(1<<i) != 0
			}
			return advice, row
		}
	}
	return drawAll, -1
}

func vpConfigFor(variant string) *VideoPokerVariantConfig {
	switch variant {
	case "deuceswild":
		return DeucesWildConfig()
	case "jokerpoker":
		return JokerPokerConfig()
	default:
		return JacksOrBetterConfig()
	}
}

func vpRank(c *Card) int {
	if v := c.GetValue(); v != 1 {
		return v
	}
	return 14
}

// vpMadeCategory は 5 枚全体の役区分を返す (該当なしは -1)。
func vpMadeCategory(cfg *VideoPokerVariantConfig, hand []*Card) int {
	rank, usedWilds := evalFiveCardHand(hand), false
	if cfg.IsWild != nil {
		rank, usedWilds = evalWildHand(hand, cfg.IsWild)
	}
	switch rank {
	case PokerHandRoyalFlush:
		if usedWilds {
			return vpCatWildRoyal
		}
		return vpCatNaturalRoyal
	case PokerHandFiveOfAKind:
		return vpCatFiveOfAKind
	case PokerHandStraightFlush:
		return vpCatStraightFlush
	case PokerHandFullHouse:
		return vpCatFullHouse
	case PokerHandFlush:
		return vpCatFlush
	case PokerHandStraight:
		return vpCatStraight
	default:
		return -1
	}
}

// vpBestMask は行 r を満たす候補から 1 つを決定的に選ぶ。
// 高位札が多い方 → ストレートの完成ランクが多い方 → 最高ランクが高い方 → マスク値が小さい方。
// unsuitedHigh / oneHigh だけは低い札を優先する (A を残すより K-Q の方が伸びる)。
func vpBestMask(r vpRule, cards []vpCard, wildMask, cat int) (int, bool) {
	best, found := 0, false
	var bestScore [3]int
	for mask := 0; mask < 1<<len(cards); mask++ {
		if mask&wildMask != wildMask {
			continue // ワイルドは常に残す
		}
		h := vpHold{mask: mask}
		for i, c := range cards {
			if mask&(1<<i) == 0 {
				continue
			}
			if c.wild {
				h.wilds++
			} else {
				h.natural = append(h.natural, c)
			}
		}
		if r.n > 0 && h.size() != r.n {
			continue
		}
		if !vpMatches(r, h, len(cards), cat) {
			continue
		}
		score := vpScore(r, h)
		if !found || vpScoreLess(bestScore, score) {
			best, bestScore, found = mask, score, true
		}
	}
	return best, found
}

func vpScoreLess(a, b [3]int) bool {
	for i := range a {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}

func vpScore(r vpRule, h vpHold) [3]int {
	high, maxRank, sum := 0, 0, 0
	for _, c := range h.natural {
		if c.rank >= 11 {
			high++
		}
		sum += c.rank
		if c.rank > maxRank {
			maxRank = c.rank
		}
	}
	if r.kind == vpUnsuitedHigh || r.kind == vpOneHigh {
		return [3]int{-sum, 0, 0}
	}
	return [3]int{high, vpStraightOuts(h), maxRank}
}

func vpMatches(r vpRule, h vpHold, handSize, cat int) bool {
	switch r.kind {
	case vpMade:
		if h.size() != handSize {
			return false
		}
		for _, c := range r.cats {
			if c == cat {
				return true
			}
		}
		return false
	case vpFourDeuces:
		return h.wilds == 4 && len(h.natural) == 0
	case vpWildsOnly:
		return h.wilds > 0 && len(h.natural) == 0
	case vpWildPlusOne:
		return h.wilds > 0 && len(h.natural) == 1 && h.natural[0].rank >= r.minRank && h.natural[0].rank <= r.maxRank
	case vpQuads, vpTrips:
		return len(h.natural) > 0 && vpSameRank(h.natural)
	case vpTwoPair:
		counts := vpRankCounts(h.natural)
		return h.wilds == 0 && len(counts) == 2 && vpAllCountsAre(counts, 2)
	case vpHighPair:
		if h.wilds == 1 && len(h.natural) == 1 {
			return h.natural[0].rank >= r.minRank
		}
		return h.wilds == 0 && vpSameRank(h.natural) && h.natural[0].rank >= r.minRank
	case vpLowPair:
		return h.wilds == 0 && vpSameRank(h.natural) && h.natural[0].rank < r.minRank
	case vpOnePair:
		return h.wilds == 0 && vpSameRank(h.natural)
	case vpRoyalDraw:
		if r.maxRank > 0 && vpCountHigh(h.natural, r.maxRank+1) > 0 {
			return false
		}
		if r.needHigh > 0 && vpCountHigh(h.natural, r.needHigh) == 0 {
			return false
		}
		return vpSuited(h.natural) && vpDistinct(h.natural) && vpAllRoyalRanks(h.natural) && vpLowestRank(h.natural) >= r.minRank
	case vpSFDraw:
		if !vpSuited(h.natural) || !vpDistinct(h.natural) || vpAllRoyalRanks(h.natural) {
			return false
		}
		if !vpFitsWindow(h.natural) || vpLowestRank(h.natural) < r.minRank {
			return false
		}
		if r.maxSpan > 0 && vpSpan(h.natural) > r.maxSpan {
			return false
		}
		return !r.consecutive || vpConsecutive(h.natural)
	case vpFlushDraw:
		if r.needHigh > 0 && vpCountHigh(h.natural, r.needHigh) == 0 {
			return false
		}
		return vpSuited(h.natural)
	case vpOutside:
		return vpDistinct(h.natural) && vpStraightOuts(h) >= 2
	case vpInside:
		if r.noAceLow && vpLowestRank(h.natural) == 1 {
			return false
		}
		return vpDistinct(h.natural) && vpStraightOuts(h) == 1 && vpCountHigh(h.natural, r.minRank) >= 3
	case vpSuitedHigh:
		return h.wilds == 0 && vpSuited(h.natural) && vpCountHigh(h.natural, r.minRank) == 2
	case vpSuitedTenHigh:
		if h.wilds != 0 || !vpSuited(h.natural) {
			return false
		}
		lo, hi := h.natural[0].rank, h.natural[1].rank
		if lo > hi {
			lo, hi = hi, lo
		}
		return lo == 10 && hi >= 11 && hi <= 13
	case vpUnsuitedHigh:
		return h.wilds == 0 && vpDistinct(h.natural) && !vpSuited(h.natural) && vpCountHigh(h.natural, r.minRank) == 2
	case vpOneHigh:
		return h.wilds == 0 && vpCountHigh(h.natural, r.minRank) == 1
	}
	return false
}

func vpRankCounts(cards []vpCard) map[int]int {
	counts := make(map[int]int, len(cards))
	for _, c := range cards {
		counts[c.rank]++
	}
	return counts
}

func vpAllCountsAre(counts map[int]int, n int) bool {
	for _, c := range counts {
		if c != n {
			return false
		}
	}
	return true
}

func vpSameRank(cards []vpCard) bool {
	for _, c := range cards {
		if c.rank != cards[0].rank {
			return false
		}
	}
	return true
}

func vpSuited(cards []vpCard) bool {
	for _, c := range cards {
		if c.suit != cards[0].suit {
			return false
		}
	}
	return true
}

func vpDistinct(cards []vpCard) bool {
	return len(vpRankCounts(cards)) == len(cards)
}

func vpAllRoyalRanks(cards []vpCard) bool {
	for _, c := range cards {
		if c.rank < 10 {
			return false
		}
	}
	return true
}

func vpCountHigh(cards []vpCard, minRank int) int {
	n := 0
	for _, c := range cards {
		if c.rank >= minRank {
			n++
		}
	}
	return n
}

// vpLowestRank は最低ランク。A は他の札が全部 5 以下なら 1 (A-5 の窓) として数える。
func vpLowestRank(cards []vpCard) int {
	lo, aceLow := 15, true
	for _, c := range cards {
		if c.rank != 14 && c.rank > 5 {
			aceLow = false
		}
	}
	for _, c := range cards {
		r := c.rank
		if r == 14 && aceLow {
			r = 1
		}
		lo = min(lo, r)
	}
	return lo
}

// vpSpan は自然札の最高と最低の差 (A は vpLowestRank と同じく A-5 の窓では 1)。
func vpSpan(cards []vpCard) int {
	lo, hi := vpLowestRank(cards), 0
	for _, c := range cards {
		r := c.rank
		if r == 14 && lo == 1 {
			r = 1
		}
		hi = max(hi, r)
	}
	return hi - lo
}

func vpConsecutive(cards []vpCard) bool {
	if len(cards) == 0 {
		return true
	}
	lo, hi := 15, 0
	for _, c := range cards {
		lo, hi = min(lo, c.rank), max(hi, c.rank)
	}
	return hi-lo == len(cards)-1
}

// vpFitsWindow は自然札 (ランクが全部異なる前提) が 5 ランクの窓 (A-5 から 10-A) に収まるか。
func vpFitsWindow(cards []vpCard) bool {
	for lo := 1; lo <= 10; lo++ {
		if vpInWindow(cards, lo) {
			return true
		}
	}
	return false
}

func vpInWindow(cards []vpCard, lo int) bool {
	for _, c := range cards {
		r := c.rank
		if r == 14 && lo == 1 {
			r = 1
		}
		if r < lo || r > lo+4 {
			return false
		}
	}
	return true
}

// vpStraightOuts は 4 枚ホールドにあと 1 ランク足してストレートになるランクの数。
// 4 枚でないホールドや、ランクが重なるホールドは 0。
func vpStraightOuts(h vpHold) int {
	if h.size() != 4 || !vpDistinct(h.natural) {
		return 0
	}
	outs := 0
	for rank := 2; rank <= 14; rank++ {
		dup := false
		for _, c := range h.natural {
			if c.rank == rank {
				dup = true
			}
		}
		if dup {
			continue
		}
		with := append(append([]vpCard(nil), h.natural...), vpCard{rank: rank})
		if vpFitsWindow(with) {
			outs++
		}
	}
	return outs
}
