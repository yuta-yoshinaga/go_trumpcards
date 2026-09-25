//go:build test

package domain

import (
	"math/rand"
	"testing"
)

const indianRummyReferenceSearchCap = 2000000

var indianRummyReferenceCapped bool

func indianRummyReferenceIsWild(card *Card, wildRank int) bool {
	if card == nil {
		return false
	}
	if card.GetDesign() == CardDesignJoker {
		return true
	}
	return wildRank != 0 && card.GetValue() == wildRank
}

// indianRummyReferenceCardPoints デッドウッド点（ワイルド=0、A=10、2-9=face、10/J/Q/K=10）。
func indianRummyReferenceCardPoints(card *Card, wildRank int) int {
	if indianRummyReferenceIsWild(card, wildRank) {
		return 0
	}
	v := card.GetValue()
	if v == 1 { // Ace
		return 10
	}
	if v >= 10 { // 10, J, Q, K
		return 10
	}
	return v
}

// --- Meld generation (joker-aware) ---

// indianRummyReferenceMeld は候補メルド。idx は cards スライス内のインデックス集合。
type indianRummyReferenceMeld struct {
	idx  []int
	seq  bool // シーケンス（ラン）か
	pure bool // ピュアシーケンス（ワイルド未使用）か
}

// indianRummyReferenceGenerateMelds cards から有効なセット／シーケンス候補を全列挙する。
func indianRummyReferenceGenerateMelds(cards []*Card, wildRank int) []indianRummyReferenceMeld {
	wildIdxs := make([]int, 0)
	for i, c := range cards {
		if indianRummyReferenceIsWild(c, wildRank) {
			wildIdxs = append(wildIdxs, i)
		}
	}
	melds := indianRummyReferenceGenerateSets(cards, wildRank, wildIdxs)
	melds = append(melds, indianRummyReferenceGenerateRuns(cards, wildRank, wildIdxs)...)
	return melds
}

// indianRummyReferenceDistinctSuitCombos idxs から k 枚を、全てスートが異なるように選ぶ組み合わせを返す。
func indianRummyReferenceDistinctSuitCombos(idxs []int, cards []*Card, k int) [][]int {
	bySuit := make(map[int][]int)
	order := make([]int, 0)
	for _, i := range idxs {
		s := cards[i].GetDesign()
		if _, ok := bySuit[s]; !ok {
			order = append(order, s)
		}
		bySuit[s] = append(bySuit[s], i)
	}
	res := make([][]int, 0)
	var rec func(pos int, cur []int)
	rec = func(pos int, cur []int) {
		if len(cur) == k {
			res = append(res, append([]int(nil), cur...))
			return
		}
		if pos >= len(order) || len(order)-pos < k-len(cur) {
			return
		}
		// このスートを使わない
		rec(pos+1, cur)
		// このスートの 1 枚を使う
		for _, i := range bySuit[order[pos]] {
			rec(pos+1, append(cur, i))
		}
	}
	rec(0, nil)
	return res
}

// indianRummyReferenceGenerateSets セット候補（同ランク別スート 3-4 枚、ワイルド最大 1 枚）を列挙する。
func indianRummyReferenceGenerateSets(cards []*Card, wildRank int, wildIdxs []int) []indianRummyReferenceMeld {
	byRank := make(map[int][]int)
	for i, c := range cards {
		if indianRummyReferenceIsWild(c, wildRank) {
			continue
		}
		byRank[c.GetValue()] = append(byRank[c.GetValue()], i)
	}
	melds := make([]indianRummyReferenceMeld, 0)
	for _, idxs := range byRank {
		// ピュアセット（ワイルドなし）3 枚・4 枚
		for _, k := range []int{3, 4} {
			for _, combo := range indianRummyReferenceDistinctSuitCombos(idxs, cards, k) {
				melds = append(melds, indianRummyReferenceMeld{idx: combo, seq: false, pure: false})
			}
		}
		// ワイルド 1 枚を含むセット（合計 3 枚・4 枚）
		for _, total := range []int{3, 4} {
			naturals := total - 1
			for _, combo := range indianRummyReferenceDistinctSuitCombos(idxs, cards, naturals) {
				for _, w := range wildIdxs {
					m := append(append([]int(nil), combo...), w)
					melds = append(melds, indianRummyReferenceMeld{idx: m, seq: false, pure: false})
				}
			}
		}
	}
	return melds
}

// indianRummyReferenceGenerateRuns シーケンス候補（同スート連続 3+ 枚、ワイルド最大 1 枚）を列挙する。
// Ace は low(A-2-3) / high(Q-K-A) の双方を許容する。
func indianRummyReferenceGenerateRuns(cards []*Card, wildRank int, wildIdxs []int) []indianRummyReferenceMeld {
	bySuit := make(map[int]map[int][]int)
	for i, c := range cards {
		if indianRummyReferenceIsWild(c, wildRank) {
			continue
		}
		s := c.GetDesign()
		if bySuit[s] == nil {
			bySuit[s] = make(map[int][]int)
		}
		v := c.GetValue()
		bySuit[s][v] = append(bySuit[s][v], i)
	}
	melds := make([]indianRummyReferenceMeld, 0)
	for _, byVal := range bySuit {
		for start := 1; start <= 13; start++ {
			for length := IndianRummySeqMin; start+length-1 <= 14; length++ {
				m, ok := indianRummyReferenceBuildRunWindow(byVal, start, length, wildIdxs)
				if ok {
					melds = append(melds, m...)
				}
			}
		}
	}
	return melds
}

// indianRummyReferenceBuildRunWindow 単一スートの value→idxs から窓 [start, start+length-1] のラン候補を作る。
func indianRummyReferenceBuildRunWindow(byVal map[int][]int, start, length int, wildIdxs []int) ([]indianRummyReferenceMeld, bool) {
	present := make([][]int, 0, length)
	missing := 0
	seen := make(map[int]bool)
	for v := start; v < start+length; v++ {
		lv := v
		if lv == 14 {
			lv = 1 // Ace-high
		}
		if seen[lv] {
			return nil, false // 同じランクを二度参照する窓（ラップアラウンド）は不可
		}
		seen[lv] = true
		opts := byVal[lv]
		if len(opts) == 0 {
			missing++
			present = append(present, nil)
		} else {
			present = append(present, opts)
		}
	}
	switch {
	case missing == 0:
		combos := indianRummyReferenceCartesian(present)
		out := make([]indianRummyReferenceMeld, 0, len(combos))
		for _, combo := range combos {
			out = append(out, indianRummyReferenceMeld{idx: combo, seq: true, pure: true})
		}
		return out, true
	case missing == 1 && len(wildIdxs) > 0:
		base := make([][]int, 0, length-1)
		for _, opts := range present {
			if len(opts) > 0 {
				base = append(base, opts)
			}
		}
		combos := indianRummyReferenceCartesian(base)
		out := make([]indianRummyReferenceMeld, 0, len(combos)*len(wildIdxs))
		for _, combo := range combos {
			for _, w := range wildIdxs {
				m := append(append([]int(nil), combo...), w)
				out = append(out, indianRummyReferenceMeld{idx: m, seq: true, pure: false})
			}
		}
		return out, true
	default:
		return nil, false
	}
}

// indianRummyReferenceCartesian 各値の候補インデックスから 1 つずつ選ぶ直積を返す（爆発防止に上限あり）。
func indianRummyReferenceCartesian(lists [][]int) [][]int {
	const maxCartesian = 256
	res := [][]int{{}}
	for _, opts := range lists {
		if len(opts) == 0 {
			continue
		}
		next := make([][]int, 0, len(res)*len(opts))
		for _, prefix := range res {
			for _, o := range opts {
				next = append(next, append(append([]int(nil), prefix...), o))
			}
		}
		if len(next) > maxCartesian {
			next = next[:maxCartesian]
		}
		res = next
	}
	return res
}

// --- Declaration / deadwood search ---

// indianRummyReferenceCovering 各カードインデックスを覆うメルドのインデックス一覧を返す。
func indianRummyReferenceCovering(n int, melds []indianRummyReferenceMeld) [][]int {
	covering := make([][]int, n)
	for mi, m := range melds {
		for _, ci := range m.idx {
			covering[ci] = append(covering[ci], mi)
		}
	}
	return covering
}

// indianRummyReferenceValidateDeclaration cards（13 枚）が有効宣言か。
// 全カードがメルドに収まり、シーケンスが 2 つ以上・うち 1 つ以上がピュアであれば true。
func indianRummyReferenceValidateDeclaration(cards []*Card, wildRank int) bool {
	n := len(cards)
	if n != IndianRummyHandSize {
		return false
	}
	melds := indianRummyReferenceGenerateMelds(cards, wildRank)
	covering := indianRummyReferenceCovering(n, melds)
	decided := make([]bool, n)
	iter := 0
	indianRummyReferenceCapped = false

	var dfs func(seq, pure int) bool
	dfs = func(seq, pure int) bool {
		iter++
		if iter > indianRummyReferenceSearchCap {
			indianRummyReferenceCapped = true
			return false
		}
		i := indianRummyReferenceFirstUndecided(decided)
		if i == -1 {
			return seq >= 2 && pure >= 1
		}
		for _, mi := range covering[i] {
			m := melds[mi]
			if !indianRummyReferenceAllUndecided(decided, m.idx) {
				continue
			}
			indianRummyReferenceSetDecided(decided, m.idx, true)
			si, pi := 0, 0
			if m.seq {
				si = 1
				if m.pure {
					pi = 1
				}
			}
			if dfs(seq+si, pure+pi) {
				return true
			}
			indianRummyReferenceSetDecided(decided, m.idx, false)
		}
		return false
	}
	return dfs(0, 0)
}

// indianRummyReferenceHasPureSequence cards にピュアシーケンス（ワイルド未使用の同スート連続 3+ 枚）が存在するか。
func indianRummyReferenceHasPureSequence(cards []*Card, wildRank int) bool {
	for _, m := range indianRummyReferenceGenerateMelds(cards, wildRank) {
		if m.seq && m.pure {
			return true
		}
	}
	return false
}

// IndianRummyDeadwoodScore デッドウッド採点値を返す。
// ピュアシーケンスが無ければ 80（フルキャップ）。あれば最小デッドウッド点を 80 で頭打ちにする。
func indianRummyReferenceDeadwoodScore(cards []*Card, wildRank int) int {
	if !indianRummyReferenceHasPureSequence(cards, wildRank) {
		return IndianRummyDeadwoodCap
	}
	dw := indianRummyReferenceMinDeadwood(cards, wildRank)
	if dw > IndianRummyDeadwoodCap {
		dw = IndianRummyDeadwoodCap
	}
	return dw
}

// indianRummyReferenceMinDeadwood cards を互いに素なメルドで覆ったときの最小デッドウッド点を返す。
func indianRummyReferenceMinDeadwood(cards []*Card, wildRank int) int {
	n := len(cards)
	if n == 0 {
		return 0
	}
	melds := indianRummyReferenceGenerateMelds(cards, wildRank)
	covering := indianRummyReferenceCovering(n, melds)
	points := make([]int, n)
	for i, c := range cards {
		points[i] = indianRummyReferenceCardPoints(c, wildRank)
	}
	decided := make([]bool, n)
	iter := 0
	indianRummyReferenceCapped = false

	var dfs func() int
	dfs = func() int {
		iter++
		if iter > indianRummyReferenceSearchCap {
			indianRummyReferenceCapped = true
			s := 0
			for k := 0; k < n; k++ {
				if !decided[k] {
					s += points[k]
				}
			}
			return s
		}
		i := indianRummyReferenceFirstUndecided(decided)
		if i == -1 {
			return 0
		}
		// 選択肢 A: カード i をデッドウッドにする
		decided[i] = true
		best := points[i] + dfs()
		decided[i] = false
		// 選択肢 B: i を覆うメルドを使う
		for _, mi := range covering[i] {
			m := melds[mi]
			if !indianRummyReferenceAllUndecided(decided, m.idx) {
				continue
			}
			indianRummyReferenceSetDecided(decided, m.idx, true)
			c := dfs()
			indianRummyReferenceSetDecided(decided, m.idx, false)
			if c < best {
				best = c
			}
		}
		return best
	}
	return dfs()
}

func indianRummyReferenceFirstUndecided(decided []bool) int {
	for i := 0; i < len(decided); i++ {
		if !decided[i] {
			return i
		}
	}
	return -1
}

func indianRummyReferenceAllUndecided(decided []bool, idx []int) bool {
	for _, ci := range idx {
		if decided[ci] {
			return false
		}
	}
	return true
}

func indianRummyReferenceSetDecided(decided []bool, idx []int, v bool) {
	for _, ci := range idx {
		decided[ci] = v
	}
}

func indianRummyPhysicalDeck() []*Card {
	d := make([]*Card, 0, 108)
	for cp := 0; cp < 2; cp++ {
		for s := 1; s <= 4; s++ {
			for v := 1; v <= 13; v++ {
				d = append(d, NewCard(s, v, false))
			}
		}
	}
	for i := 0; i < 4; i++ {
		d = append(d, NewCard(CardDesignJoker, CardValueJoker, false))
	}
	return d
}
func TestIndianRummyMeldEngineMatchesLegacyReference(t *testing.T) {
	rng := rand.New(rand.NewSource(805599))
	deck := indianRummyPhysicalDeck()
	matched, truth, falsity, capped := 0, 0, 0, 0
	check := func(cards []*Card, wild int) {
		indianRummyReferenceCapped = false
		old := indianRummyReferenceMinDeadwood(cards, wild)
		capMin := indianRummyReferenceCapped
		oldPure := indianRummyReferenceHasPureSequence(cards, wild)
		oldScore := indianRummyReferenceDeadwoodScore(cards, wild)
		indianRummyReferenceCapped = false
		oldDecl := indianRummyReferenceValidateDeclaration(cards, wild)
		capDecl := indianRummyReferenceCapped
		newMin, _ := rummyMinDeadwood(cards, indianRummyRules(wild))
		gotPure := IndianRummyHasPureSequence(cards, wild)
		gotScore := IndianRummyDeadwoodScore(cards, wild)
		gotDecl := IndianRummyValidateDeclaration(cards, wild)
		if capMin || capDecl {
			capped++
			if newMin > old {
				t.Fatalf("capped min got=%d ref=%d", newMin, old)
			}
			if oldDecl && !gotDecl {
				t.Fatal("capped declaration reference true but new false")
			}
		} else {
			if newMin != old || gotScore != oldScore || gotPure != oldPure || gotDecl != oldDecl {
				t.Fatalf("parity min %d/%d score %d/%d pure %v/%v decl %v/%v", newMin, old, gotScore, oldScore, gotPure, oldPure, gotDecl, oldDecl)
			}
			matched++
		}
		if gotDecl {
			truth++
		} else {
			falsity++
		}
	}
	for i := 0; i < 320; i++ {
		rng.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
		n := 13
		if i%2 == 1 {
			n = 14
		}
		wild := 0
		if i%3 != 0 {
			wild = 1 + rng.Intn(13)
		}
		check(append([]*Card(nil), deck[:n]...), wild)
	}
	// Generate valid declarations from physical deck; two pure sequences, a set, and a run cover all 13 cards.
	for i := 0; i < 60; i++ {
		hand := []*Card{NewCard(1, 3, false), NewCard(1, 4, false), NewCard(1, 5, false), NewCard(2, 6, false), NewCard(2, 7, false), NewCard(2, 8, false), NewCard(3, 9, false), NewCard(4, 9, false), NewCard(1, 9, false), NewCard(4, 10, false), NewCard(4, 11, false), NewCard(4, 12, false), NewCard(4, 13, false)}
		rng.Shuffle(len(hand), func(i, j int) { hand[i], hand[j] = hand[j], hand[i] })
		check(hand, 0)
	}
	if truth == 0 || falsity == 0 || matched == 0 {
		t.Fatalf("coverage true=%d false=%d matched=%d", truth, falsity, matched)
	}
	t.Logf("hands=380 matched=%d true=%d false=%d capped=%d", matched, truth, falsity, capped)
}

func TestIndianRummyDeclarationBoundaries(t *testing.T) {
	rng := rand.New(rand.NewSource(8055001))
	sets := func() []*Card {
		return []*Card{NewCard(1, 5, false), NewCard(2, 5, false), NewCard(3, 5, false), NewCard(1, 7, false), NewCard(2, 7, false), NewCard(3, 7, false), NewCard(1, 9, false), NewCard(2, 9, false), NewCard(3, 9, false), NewCard(4, 9, false)}
	}
	cases := []struct {
		name string
		hand func() []*Card
		want bool
	}{
		{"one pure sequence and sets", func() []*Card {
			return append([]*Card{NewCard(1, 1, false), NewCard(1, 2, false), NewCard(1, 3, false)}, sets()...)
		}, false},
		{"two impure sequences and sets", func() []*Card {
			return []*Card{NewCard(1, 1, false), NewCard(1, 3, false), NewCard(4, 13, false), NewCard(2, 6, false), NewCard(2, 8, false), NewCard(3, 13, false), NewCard(1, 4, false), NewCard(2, 4, false), NewCard(3, 4, false), NewCard(1, 10, false), NewCard(2, 10, false), NewCard(3, 10, false), NewCard(4, 10, false)}
		}, false},
		{"one pure and one impure sequence and sets", func() []*Card {
			return []*Card{NewCard(1, 1, false), NewCard(1, 2, false), NewCard(1, 3, false), NewCard(2, 6, false), NewCard(2, 7, false), NewCard(4, 13, false), NewCard(1, 5, false), NewCard(2, 5, false), NewCard(3, 5, false), NewCard(1, 9, false), NewCard(2, 9, false), NewCard(3, 9, false), NewCard(4, 9, false)}
		}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for i := 0; i < 30; i++ {
				hand := tc.hand()
				rng.Shuffle(len(hand), func(i, j int) { hand[i], hand[j] = hand[j], hand[i] })
				old := indianRummyReferenceValidateDeclaration(hand, 13)
				got := IndianRummyValidateDeclaration(hand, 13)
				if old != tc.want || got != tc.want {
					t.Fatalf("sample %d: reference=%v got=%v want=%v", i, old, got, tc.want)
				}
			}
		})
	}
}
