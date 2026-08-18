//go:build !js || !wasm || extra2

package domain

import "sort"

// ティチュー特殊カード種別 (CardDesignJoker のカード値で識別)
const (
	// TichuMahjong 麻雀 (一鳥): 最弱の単体、先手権を持つ
	TichuMahjong = 1
	// TichuDog 犬: 先手をパートナーに譲る
	TichuDog = 2
	// TichuPhoenix 鳳凰: ワイルドカード (得点 -25)
	TichuPhoenix = 3
	// TichuDragon 龍: 最強の単体 (得点 +25) だがトリックは相手に渡る
	TichuDragon = 4
)

// ティチューの特殊ランク
const (
	tichuMahjongRank = 1
	tichuDragonRank  = 15
)

// TichuComboType 役種別
type TichuComboType int

// TichuComboType定数
const (
	// TichuComboInvalid 無効
	TichuComboInvalid TichuComboType = 0
	// TichuComboSingle 単体
	TichuComboSingle TichuComboType = 1
	// TichuComboPair ペア
	TichuComboPair TichuComboType = 2
	// TichuComboTriple スリーカード
	TichuComboTriple TichuComboType = 3
	// TichuComboFullHouse フルハウス (3+2)
	TichuComboFullHouse TichuComboType = 4
	// TichuComboStraight 階段 (5枚以上の連続)
	TichuComboStraight TichuComboType = 5
	// TichuComboStairs 連続ペア (2ペア以上)
	TichuComboStairs TichuComboType = 6
	// TichuComboBomb 4枚のボム
	TichuComboBomb TichuComboType = 7
	// TichuComboStraightFlush ストレートフラッシュ (同スート5枚以上の連続)
	TichuComboStraightFlush TichuComboType = 8
	// TichuComboDog 犬 (特殊リード)
	TichuComboDog TichuComboType = 9
)

// TichuCombo ティチューの役
type TichuCombo struct {
	Type          TichuComboType
	Cards         []*Card
	Rank          int  // 比較用の主ランク
	Length        int  // 階段・連続ペアの枚数
	PhoenixSingle bool // 鳳凰単体 (テーブル依存で半ランク上)
}

// tichuSpecialKind 特殊カードの種別を返す (0=通常カード)
func tichuSpecialKind(c *Card) int {
	if c == nil || c.GetDesign() != CardDesignJoker {
		return 0
	}
	return c.GetValue()
}

// tichuRank コンボ内で用いる非鳳凰カードのランク (鳳凰・犬は -1)
func tichuRank(c *Card) int {
	switch tichuSpecialKind(c) {
	case TichuMahjong:
		return tichuMahjongRank
	case TichuDragon:
		return tichuDragonRank
	case TichuPhoenix, TichuDog:
		return -1
	default:
		v := c.GetValue()
		if v == 1 { // Ace は最強
			return 14
		}
		return v
	}
}

// TichuCardStrength 手札ソート用の強さ (弱い順)
func TichuCardStrength(c *Card) int {
	switch tichuSpecialKind(c) {
	case TichuMahjong:
		return 1
	case TichuPhoenix:
		return 15
	case TichuDragon:
		return 16
	case TichuDog:
		return 17
	default:
		return tichuRank(c)
	}
}

// TichuCardPoints カードの得点を返す
func TichuCardPoints(c *Card) int {
	switch tichuSpecialKind(c) {
	case TichuDragon:
		return 25
	case TichuPhoenix:
		return -25
	}
	switch c.GetValue() {
	case 5:
		return 5
	case 10, 13:
		return 10
	default:
		return 0
	}
}

// TichuCardsPoints カード集合の合計得点
func TichuCardsPoints(cards []*Card) int {
	total := 0
	for _, c := range cards {
		total += TichuCardPoints(c)
	}
	return total
}

// tichuPhoenixCount 鳳凰の枚数
func tichuPhoenixCount(cards []*Card) int {
	n := 0
	for _, c := range cards {
		if tichuSpecialKind(c) == TichuPhoenix {
			n++
		}
	}
	return n
}

// tichuNonPhoenix 鳳凰を除いたカード
func tichuNonPhoenix(cards []*Card) []*Card {
	res := make([]*Card, 0, len(cards))
	for _, c := range cards {
		if tichuSpecialKind(c) != TichuPhoenix {
			res = append(res, c)
		}
	}
	return res
}

// tichuRankCounts 非鳳凰カードのランク別枚数
func tichuRankCounts(cards []*Card) map[int]int {
	m := make(map[int]int)
	for _, c := range cards {
		m[tichuRank(c)]++
	}
	return m
}

// tichuSortedDistinctRanks 昇順の重複なしランク列
func tichuSortedDistinctRanks(counts map[int]int) []int {
	ranks := make([]int, 0, len(counts))
	for r := range counts {
		ranks = append(ranks, r)
	}
	sort.Ints(ranks)
	return ranks
}

// ClassifyTichu 選択カードを役として分類する (無効なら nil)
func ClassifyTichu(cards []*Card) *TichuCombo {
	n := len(cards)
	if n == 0 {
		return nil
	}

	var hasDog, hasDragon bool
	for _, c := range cards {
		switch tichuSpecialKind(c) {
		case TichuDog:
			hasDog = true
		case TichuDragon:
			hasDragon = true
		}
	}
	pcount := tichuPhoenixCount(cards)

	// 犬は単体リードのみ
	if hasDog {
		if n == 1 {
			return &TichuCombo{Type: TichuComboDog, Cards: cards}
		}
		return nil
	}
	// 龍は単体のみ
	if hasDragon {
		if n == 1 {
			return &TichuCombo{Type: TichuComboSingle, Cards: cards, Rank: tichuDragonRank}
		}
		return nil
	}

	// 単体
	if n == 1 {
		if pcount == 1 {
			return &TichuCombo{Type: TichuComboSingle, Cards: cards, PhoenixSingle: true, Rank: tichuMahjongRank}
		}
		return &TichuCombo{Type: TichuComboSingle, Cards: cards, Rank: tichuRank(cards[0])}
	}

	if pcount > 1 {
		return nil // 鳳凰は1枚のみ
	}

	switch n {
	case 2:
		return tichuTryPair(cards, pcount)
	case 3:
		return tichuTryTriple(cards, pcount)
	default:
		if c := tichuTryBomb4(cards, pcount); c != nil {
			return c
		}
		if c := tichuTryStraightFlush(cards, pcount); c != nil {
			return c
		}
		if c := tichuTryStraight(cards, pcount); c != nil {
			return c
		}
		if c := tichuTryStairs(cards, pcount); c != nil {
			return c
		}
		if c := tichuTryFullHouse(cards, pcount); c != nil {
			return c
		}
		return nil
	}
}

// tichuTryPair ペア判定
func tichuTryPair(cards []*Card, pcount int) *TichuCombo {
	rest := tichuNonPhoenix(cards)
	if pcount == 1 {
		// 鳳凰 + 通常1枚
		r := tichuRank(rest[0])
		if r < tichuMahjongRank+1 { // 麻雀(1)とのペアは不可 (実質ありえないが安全側)
			return nil
		}
		return &TichuCombo{Type: TichuComboPair, Cards: cards, Rank: r}
	}
	if tichuRank(rest[0]) == tichuRank(rest[1]) && tichuRank(rest[0]) >= 2 {
		return &TichuCombo{Type: TichuComboPair, Cards: cards, Rank: tichuRank(rest[0])}
	}
	return nil
}

// tichuTryTriple スリーカード判定
func tichuTryTriple(cards []*Card, pcount int) *TichuCombo {
	rest := tichuNonPhoenix(cards)
	counts := tichuRankCounts(rest)
	if pcount == 1 {
		// 通常2枚が同ランク + 鳳凰
		if len(counts) == 1 {
			r := tichuRank(rest[0])
			if r >= 2 {
				return &TichuCombo{Type: TichuComboTriple, Cards: cards, Rank: r}
			}
		}
		return nil
	}
	if len(counts) == 1 && tichuRank(rest[0]) >= 2 {
		return &TichuCombo{Type: TichuComboTriple, Cards: cards, Rank: tichuRank(rest[0])}
	}
	return nil
}

// tichuTryBomb4 4枚ボム判定 (鳳凰不可・4枚同ランク)
func tichuTryBomb4(cards []*Card, pcount int) *TichuCombo {
	if pcount != 0 || len(cards) != 4 {
		return nil
	}
	counts := tichuRankCounts(cards)
	if len(counts) == 1 {
		r := tichuRank(cards[0])
		if r >= 2 && r <= 14 {
			return &TichuCombo{Type: TichuComboBomb, Cards: cards, Rank: r, Length: 4}
		}
	}
	return nil
}

// tichuTryStraightFlush ストレートフラッシュ判定 (鳳凰不可・同スート連続5枚以上)
func tichuTryStraightFlush(cards []*Card, pcount int) *TichuCombo {
	if pcount != 0 || len(cards) < 5 {
		return nil
	}
	design := -1
	for _, c := range cards {
		if c.GetDesign() == CardDesignJoker { // 麻雀等の特殊はスート無し
			return nil
		}
		if design == -1 {
			design = c.GetDesign()
		} else if c.GetDesign() != design {
			return nil
		}
	}
	counts := tichuRankCounts(cards)
	ranks := tichuSortedDistinctRanks(counts)
	if len(ranks) != len(cards) {
		return nil // 重複ランクあり
	}
	if ranks[len(ranks)-1]-ranks[0] != len(ranks)-1 {
		return nil // 連続でない
	}
	return &TichuCombo{Type: TichuComboStraightFlush, Cards: cards, Rank: ranks[len(ranks)-1], Length: len(cards)}
}

// tichuTryStraight 階段 (連続単体5枚以上) 判定。鳳凰は1枚まで穴埋め/延長に使用可。
func tichuTryStraight(cards []*Card, pcount int) *TichuCombo {
	if len(cards) < 5 {
		return nil
	}
	rest := tichuNonPhoenix(cards)
	counts := tichuRankCounts(rest)
	ranks := tichuSortedDistinctRanks(counts)
	if len(ranks) != len(rest) {
		return nil // 重複ランクあり (ペアが混入)
	}
	for _, r := range ranks {
		if r < 1 || r > 14 { // 龍(15)は階段に入れない
			return nil
		}
	}
	if pcount == 0 {
		if ranks[len(ranks)-1]-ranks[0] == len(ranks)-1 {
			return &TichuCombo{Type: TichuComboStraight, Cards: cards, Rank: ranks[len(ranks)-1], Length: len(cards)}
		}
		return nil
	}
	// 鳳凰1枚: 連続なら端を延長、1穴なら穴埋め
	lo, hi := ranks[0], ranks[len(ranks)-1]
	span := hi - lo
	switch span {
	case len(ranks) - 1: // 隙間なし → 端を延長
		topRank := hi
		if hi+1 <= 14 {
			topRank = hi + 1
		}
		return &TichuCombo{Type: TichuComboStraight, Cards: cards, Rank: topRank, Length: len(cards)}
	case len(ranks): // 1つ穴 → 穴埋め (最高ランクは hi)
		return &TichuCombo{Type: TichuComboStraight, Cards: cards, Rank: hi, Length: len(cards)}
	default:
		return nil
	}
}

// tichuTryStairs 連続ペア (2ペア以上・偶数枚) 判定。鳳凰は1枚まで使用可。
func tichuTryStairs(cards []*Card, pcount int) *TichuCombo {
	if len(cards) < 4 || len(cards)%2 != 0 {
		return nil
	}
	rest := tichuNonPhoenix(cards)
	counts := tichuRankCounts(rest)
	for r := range counts {
		if r < 2 || r > 14 { // 麻雀(1)/龍(15)はペアに使えない
			return nil
		}
	}
	ranks := tichuSortedDistinctRanks(counts)
	if ranks[len(ranks)-1]-ranks[0] != len(ranks)-1 {
		return nil // ランクが連続でない
	}
	singles := 0
	for _, r := range ranks {
		switch counts[r] {
		case 2:
		case 1:
			singles++
		default:
			return nil
		}
	}
	hi := ranks[len(ranks)-1]
	if pcount == 0 {
		if singles == 0 {
			return &TichuCombo{Type: TichuComboStairs, Cards: cards, Rank: hi, Length: len(cards)}
		}
		return nil
	}
	// 鳳凰1枚: ちょうど1ランクが単独でなければならない
	if singles == 1 {
		return &TichuCombo{Type: TichuComboStairs, Cards: cards, Rank: hi, Length: len(cards)}
	}
	return nil
}

// tichuTryFullHouse フルハウス (3+2) 判定。鳳凰は1枚まで使用可。
func tichuTryFullHouse(cards []*Card, pcount int) *TichuCombo {
	if len(cards) != 5 {
		return nil
	}
	rest := tichuNonPhoenix(cards)
	counts := tichuRankCounts(rest)
	for r := range counts {
		if r < 2 || r > 14 {
			return nil
		}
	}
	if pcount == 0 {
		var trip, pair int
		for r, c := range counts {
			switch c {
			case 3:
				trip = r
			case 2:
				pair = r
			default:
				return nil
			}
		}
		if trip != 0 && pair != 0 {
			return &TichuCombo{Type: TichuComboFullHouse, Cards: cards, Rank: trip}
		}
		return nil
	}
	// 鳳凰1枚 (通常4枚)
	ranks := tichuSortedDistinctRanks(counts)
	switch len(ranks) {
	case 2:
		r0, r1 := ranks[0], ranks[1]
		c0, c1 := counts[r0], counts[r1]
		if c0 == 3 && c1 == 1 { // 鳳凰でペア完成、トリップは r0
			return &TichuCombo{Type: TichuComboFullHouse, Cards: cards, Rank: r0}
		}
		if c0 == 1 && c1 == 3 {
			return &TichuCombo{Type: TichuComboFullHouse, Cards: cards, Rank: r1}
		}
		if c0 == 2 && c1 == 2 { // 鳳凰で高い方をトリップに
			return &TichuCombo{Type: TichuComboFullHouse, Cards: cards, Rank: r1}
		}
		return nil
	default:
		return nil
	}
}

// tichuIsBomb ボム系か
func tichuIsBomb(combo *TichuCombo) bool {
	return combo != nil && (combo.Type == TichuComboBomb || combo.Type == TichuComboStraightFlush)
}

// tichuSingleValueX2 単体役の比較値 (×2スケールで鳳凰の半ランクを表現)
func tichuSingleValueX2(combo *TichuCombo) int {
	if combo.PhoenixSingle {
		return combo.Rank*2 + 1
	}
	return combo.Rank * 2
}

// TichuCanBeat cand が table を上回るか判定する。
// table が nil の場合 (リード) は呼び出し側で常に許可するため、ここでは table != nil を前提とする。
func TichuCanBeat(cand, table *TichuCombo) bool {
	if cand == nil || table == nil {
		return false
	}
	if cand.Type == TichuComboDog {
		return false // 犬はリード専用
	}

	candBomb := tichuIsBomb(cand)
	tableBomb := tichuIsBomb(table)

	if candBomb && !tableBomb {
		return true // ボムは非ボムを必ず上回る
	}
	if !candBomb && tableBomb {
		return false
	}
	if candBomb && tableBomb {
		return tichuCompareBombs(cand, table)
	}

	// 非ボム同士: 同じ役種・同じ枚数で高ランク
	if cand.Type != table.Type {
		return false
	}
	switch cand.Type {
	case TichuComboSingle:
		// 鳳凰は龍を超えられない
		if cand.PhoenixSingle && !table.PhoenixSingle && table.Rank == tichuDragonRank {
			return false
		}
		// 龍は最強、鳳凰はテーブル依存
		return tichuSingleValueX2(cand) > tichuSingleValueX2(table)
	case TichuComboStraight, TichuComboStairs:
		if cand.Length != table.Length {
			return false
		}
		return cand.Rank > table.Rank
	default:
		return cand.Rank > table.Rank
	}
}

// tichuCompareBombs ボム同士の比較
func tichuCompareBombs(cand, table *TichuCombo) bool {
	candSF := cand.Type == TichuComboStraightFlush
	tableSF := table.Type == TichuComboStraightFlush
	if candSF && !tableSF {
		return true // ストレートフラッシュ > 4枚ボム
	}
	if !candSF && tableSF {
		return false
	}
	if candSF && tableSF {
		if cand.Length != table.Length {
			return cand.Length > table.Length // 長い方が強い
		}
		return cand.Rank > table.Rank
	}
	// 4枚ボム同士
	return cand.Rank > table.Rank
}

// TichuBombIndices は手札のうちボム (同ランク4枚、または同スート5枚以上の連続) を
// 構成できるカードの位置を返す。
//
// **判定そのものは tichuTryBomb4 / tichuTryStraightFlush が持っている。**ここは
// 「どの組がボムになるか」を数え上げるだけで、成立の可否は必ずその2つに訊く。
// 数え上げ側でルールを書き直すと、画面が「ボム」と言った組を出そうとして
// 弾かれる状態が作れる (#5635)。
func TichuBombIndices(cards []*Card) []int {
	marked := make(map[int]bool, len(cards))

	// 同ランク4枚。特殊カード (ジョーカー扱い) は参加しない。
	byRank := make(map[int][]int, len(cards))
	for i, c := range cards {
		if c == nil || c.GetDesign() == CardDesignJoker {
			continue
		}
		byRank[tichuRank(c)] = append(byRank[tichuRank(c)], i)
	}
	for _, idxs := range byRank {
		if len(idxs) < 4 {
			continue
		}
		group := make([]*Card, 0, 4)
		for _, i := range idxs[:4] {
			group = append(group, cards[i])
		}
		if tichuTryBomb4(group, 0) == nil {
			continue
		}
		for _, i := range idxs[:4] {
			marked[i] = true
		}
	}

	// 同スートの連続。長い方から試し、成立した並びを印にする。
	bySuit := make(map[int][]int, 4)
	for i, c := range cards {
		if c == nil || c.GetDesign() == CardDesignJoker {
			continue
		}
		bySuit[c.GetDesign()] = append(bySuit[c.GetDesign()], i)
	}
	for _, idxs := range bySuit {
		sort.Slice(idxs, func(a, b int) bool { return tichuRank(cards[idxs[a]]) < tichuRank(cards[idxs[b]]) })
		for start := range idxs {
			for end := len(idxs); end > start; end-- {
				run := idxs[start:end]
				if len(run) < 5 {
					break
				}
				group := make([]*Card, 0, len(run))
				for _, i := range run {
					group = append(group, cards[i])
				}
				if tichuTryStraightFlush(group, 0) == nil {
					continue
				}
				for _, i := range run {
					marked[i] = true
				}
				break
			}
		}
	}

	out := make([]int, 0, len(marked))
	for i := range cards {
		if marked[i] {
			out = append(out, i)
		}
	}
	return out
}
