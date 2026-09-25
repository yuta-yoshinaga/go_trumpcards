//go:build test

package domain

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

const marriageSearchCap = 2000000
const marriagePureSequenceSearchCap = 10000

// --- Meld generation (joker-aware) ---

// marriageMeld は候補メルド。idx は cards スライス内のインデックス集合。
type marriageMeld struct {
	idx  []int
	seq  bool // シーケンス（ラン）か
	pure bool // ピュアシーケンス（ワイルド未使用）か
}

// marriageGenerateMelds cards から有効なセット／シーケンス候補を全列挙する。
func marriageGenerateMelds(cards []*Card, wildRank int) []marriageMeld {
	wildIdxs := make([]int, 0)
	for i, c := range cards {
		if marriageIsWild(c, wildRank) {
			wildIdxs = append(wildIdxs, i)
		}
	}
	melds := marriageGenerateSets(cards, wildRank, wildIdxs)
	melds = append(melds, marriageGenerateRuns(cards, wildRank, wildIdxs)...)
	return melds
}

// marriageDistinctSuitCombos idxs から k 枚を、全てスートが異なるように選ぶ組み合わせを返す。
func marriageDistinctSuitCombos(idxs []int, cards []*Card, k int) [][]int {
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

// marriageGenerateSets セット候補（同ランク別スート 3-4 枚、ワイルド最大 1 枚）を列挙する。
func marriageGenerateSets(cards []*Card, wildRank int, wildIdxs []int) []marriageMeld {
	byRank := make(map[int][]int)
	for i, c := range cards {
		if marriageIsWild(c, wildRank) {
			continue
		}
		byRank[c.GetValue()] = append(byRank[c.GetValue()], i)
	}
	melds := make([]marriageMeld, 0)
	for _, idxs := range byRank {
		// ピュアセット（ワイルドなし）3 枚・4 枚
		for _, k := range []int{3, 4} {
			for _, combo := range marriageDistinctSuitCombos(idxs, cards, k) {
				melds = append(melds, marriageMeld{idx: combo, seq: false, pure: false})
			}
		}
		// ワイルド 1 枚を含むセット（合計 3 枚・4 枚）
		for _, total := range []int{3, 4} {
			naturals := total - 1
			for _, combo := range marriageDistinctSuitCombos(idxs, cards, naturals) {
				for _, w := range wildIdxs {
					m := append(append([]int(nil), combo...), w)
					melds = append(melds, marriageMeld{idx: m, seq: false, pure: false})
				}
			}
		}
	}
	return melds
}

// marriageGenerateRuns シーケンス候補（同スート連続 3+ 枚、ワイルド最大 1 枚）を列挙する。
// Ace は low(A-2-3) / high(Q-K-A) の双方を許容する。
func marriageGenerateRuns(cards []*Card, wildRank int, wildIdxs []int) []marriageMeld {
	bySuit := make(map[int]map[int][]int)
	for i, c := range cards {
		if marriageIsWild(c, wildRank) {
			continue
		}
		s := c.GetDesign()
		if bySuit[s] == nil {
			bySuit[s] = make(map[int][]int)
		}
		v := c.GetValue()
		bySuit[s][v] = append(bySuit[s][v], i)
	}
	melds := make([]marriageMeld, 0)
	for _, byVal := range bySuit {
		for start := 1; start <= 13; start++ {
			for length := MarriageSeqMin; start+length-1 <= 14; length++ {
				m, ok := marriageBuildRunWindow(byVal, start, length, wildIdxs)
				if ok {
					melds = append(melds, m...)
				}
			}
		}
	}
	return melds
}

// marriageBuildRunWindow 単一スートの value→idxs から窓 [start, start+length-1] のラン候補を作る。
func marriageBuildRunWindow(byVal map[int][]int, start, length int, wildIdxs []int) ([]marriageMeld, bool) {
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
		combos := marriageCartesian(present)
		out := make([]marriageMeld, 0, len(combos))
		for _, combo := range combos {
			out = append(out, marriageMeld{idx: combo, seq: true, pure: true})
		}
		return out, true
	case missing == 1 && len(wildIdxs) > 0:
		base := make([][]int, 0, length-1)
		for _, opts := range present {
			if len(opts) > 0 {
				base = append(base, opts)
			}
		}
		combos := marriageCartesian(base)
		out := make([]marriageMeld, 0, len(combos)*len(wildIdxs))
		for _, combo := range combos {
			for _, w := range wildIdxs {
				m := append(append([]int(nil), combo...), w)
				out = append(out, marriageMeld{idx: m, seq: true, pure: false})
			}
		}
		return out, true
	default:
		return nil, false
	}
}

// marriageCartesian 各値の候補インデックスから 1 つずつ選ぶ直積を返す（爆発防止に上限あり）。
func marriageCartesian(lists [][]int) [][]int {
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

func marriageCovering(n int, melds []marriageMeld) [][]int {
	covering := make([][]int, n)
	for mi, m := range melds {
		for _, ci := range m.idx {
			covering[ci] = append(covering[ci], mi)
		}
	}
	return covering
}

func marriageFirstUndecided(decided []bool) int {
	for i, v := range decided {
		if !v {
			return i
		}
	}
	return -1
}
func marriageAllUndecided(decided []bool, idx []int) bool {
	for _, i := range idx {
		if decided[i] {
			return false
		}
	}
	return true
}
func marriageSetDecided(decided []bool, idx []int, v bool) {
	for _, i := range idx {
		decided[i] = v
	}
}

func marriageRandomPhysicalHand(rng *rand.Rand, n int) []*Card {
	deck := make([]*Card, 0, 162)
	for copy := 0; copy < 3; copy++ {
		for suit := 1; suit <= 4; suit++ {
			for rank := 1; rank <= 13; rank++ {
				deck = append(deck, NewCard(suit, rank, false))
			}
		}
	}
	for i := 0; i < 6; i++ {
		deck = append(deck, NewCard(CardDesignJoker, CardValueJoker, false))
	}
	rng.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck[:n]
}

func marriageValidateDeclarationReference(cards []*Card, wildRank int) (bool, int) {
	n := len(cards)
	if n != MarriageHandSize {
		return false, 0
	}
	melds := marriageGenerateMelds(cards, wildRank)
	covering := marriageCovering(n, melds)
	decided := make([]bool, n)
	iter := 0
	var dfs func(int) bool
	dfs = func(pure int) bool {
		iter++
		if iter > marriageSearchCap {
			return false
		}
		i := marriageFirstUndecided(decided)
		if i < 0 {
			return pure >= 3
		}
		for _, mi := range covering[i] {
			m := melds[mi]
			if !marriageAllUndecided(decided, m.idx) {
				continue
			}
			marriageSetDecided(decided, m.idx, true)
			add := 0
			if m.seq && m.pure {
				add = 1
			}
			if dfs(pure + add) {
				return true
			}
			marriageSetDecided(decided, m.idx, false)
		}
		return false
	}
	return dfs(0), iter
}

func TestMarriageValidateDeclarationMemoMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(805199))
	trueCount, falseCount, capped := 0, 0, 0
	for i := 0; i < 120; i++ {
		cards := marriageValidDeclarationHand()
		rng.Shuffle(len(cards), func(i, j int) { cards[i], cards[j] = cards[j], cards[i] })
		old, iters := marriageValidateDeclarationReference(cards, 0)
		got := MarriageValidateDeclaration(cards, 0)
		if iters <= marriageSearchCap && got != old {
			t.Fatalf("valid sample %d: got %v want %v", i, got, old)
		}
		if iters > marriageSearchCap {
			capped++
		}
		if got {
			trueCount++
		} else {
			falseCount++
		}
	}
	// Random hands exercise false outcomes and retain old/new parity when the old search finishes.
	for i := 0; i < 200; i++ {
		cards := make([]*Card, MarriageHandSize)
		for j := range cards {
			cards[j] = NewCard(1+rng.Intn(4), 1+rng.Intn(13), false)
		}
		wildRank := 1 + rng.Intn(13)
		old, iters := marriageValidateDeclarationReference(cards, wildRank)
		got := MarriageValidateDeclaration(cards, wildRank)
		if iters <= marriageSearchCap && got != old {
			t.Fatalf("random sample %d: got %v want %v", i, got, old)
		}
		if iters > marriageSearchCap {
			capped++
		}
		if got {
			trueCount++
		} else {
			falseCount++
		}
	}
	if trueCount < 50 || falseCount == 0 {
		t.Fatalf("coverage: true=%d false=%d", trueCount, falseCount)
	}
	t.Logf("declaration outcomes true=%d false=%d old-search capped=%d", trueCount, falseCount, capped)
}

// marriageMinDeadwoodMemoReference is the pre-canonical memoized algorithm.
func marriageMinDeadwoodMemoReference(cards []*Card, wildRank int) int {
	n := len(cards)
	melds := marriageGenerateMelds(cards, wildRank)
	covering := marriageCovering(n, melds)
	masks := make([]uint64, len(melds))
	for i, m := range melds {
		for _, j := range m.idx {
			masks[i] |= uint64(1) << j
		}
	}
	points := make([]int, n)
	for i, c := range cards {
		points[i] = marriageCardPoints(c, wildRank)
	}
	full := (uint64(1) << n) - 1
	memo := map[uint64]int{}
	var dfs func(uint64) int
	dfs = func(mask uint64) int {
		if mask == full {
			return 0
		}
		if v, ok := memo[mask]; ok {
			return v
		}
		remaining := ^mask
		i := 0
		for remaining&(uint64(1)<<i) == 0 {
			i++
		}
		best := points[i] + dfs(mask|(uint64(1)<<i))
		for _, mi := range covering[i] {
			m := masks[mi]
			if mask&m == 0 {
				if v := dfs(mask | m); v < best {
					best = v
				}
			}
		}
		memo[mask] = best
		return best
	}
	return dfs(0)
}

func marriageHasPureSequencesReference(cards []*Card, wildRank, n int) (bool, int) {
	melds := marriageGenerateMelds(cards, wildRank)
	used := make([]bool, len(cards))
	iter := 0
	var dfs func(int, int) bool
	dfs = func(pos, count int) bool {
		iter++
		if iter > marriagePureSequenceSearchCap {
			return false
		}
		if count >= n {
			return true
		}
		for i := pos; i < len(melds); i++ {
			x := melds[i]
			if !x.seq || !x.pure || !marriageAllUndecided(used, x.idx) {
				continue
			}
			marriageSetDecided(used, x.idx, true)
			if dfs(i+1, count+1) {
				return true
			}
			marriageSetDecided(used, x.idx, false)
		}
		return false
	}
	ok := dfs(0, 0)
	return ok, iter
}

func TestMarriagePureSequenceSearchNeverMissesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(805177))
	for _, n := range []int{1, 2, 3} {
		trueCount, falseCount, capped, matched := 0, 0, 0, 0
		for sample := 0; sample < 320; sample++ {
			cards := marriageRandomPhysicalHand(rng, 21+sample%2)
			wildRank := 0
			if sample%5 == 0 {
				wildRank = 1 + rng.Intn(13)
			}
			ref, iters := marriageHasPureSequencesReference(cards, wildRank, n)
			got := MarriageHasPureSequences(cards, wildRank, n)
			if iters <= marriagePureSequenceSearchCap {
				matched++
				if ref != got {
					t.Fatalf("n=%d sample=%d got=%v reference=%v", n, sample, got, ref)
				}
			} else {
				capped++
				if ref && !got {
					t.Fatalf("n=%d sample=%d missed capped reference true", n, sample)
				}
			}
			if got {
				trueCount++
			} else {
				falseCount++
			}
		}
		if trueCount == 0 || falseCount == 0 || matched == 0 {
			t.Fatalf("n=%d outcomes true=%d false=%d matched=%d", n, trueCount, falseCount, matched)
		}
		t.Logf("pure sequences n=%d: true=%d false=%d matched=%d capped=%d", n, trueCount, falseCount, matched, capped)
	}
}

func TestMarriageMinDeadwoodCanonicalMatchesMemoReference(t *testing.T) {
	rng := rand.New(rand.NewSource(805122))
	matched, improved := 0, 0
	var example string
	for sample := 0; sample < 320; sample++ {
		n := 21 + sample%2
		cards := marriageRandomPhysicalHand(rng, n)
		wildRank := 0
		if sample%10 == 0 {
			wildRank = 1 + rng.Intn(13)
		}
		want := marriageMinDeadwoodMemoReference(cards, wildRank)
		got := marriageMinDeadwood(cards, wildRank)
		if got > want {
			t.Fatalf("sample %d: got %d exceeds reference %d", sample, got, want)
		}
		if got == want {
			matched++
		} else {
			improved++
			if example == "" {
				example = fmt.Sprintf("sample %d: new=%d reference=%d", sample, got, want)
			}
		}
	}
	if matched == 0 {
		t.Fatalf("comparison had no matching cases: matched=%d improved=%d", matched, improved)
	}
	t.Logf("new/reference deadwood: matched=%d improved=%d; example %s", matched, improved, example)
}

func TestMarriageMinDeadwoodCanonicalWorstCaseMemoStates(t *testing.T) {
	// Identical physical copies maximize the symmetry the canonical order removes.
	cards := make([]*Card, 22)
	for i := range cards {
		cards[i] = NewCard(1+i%4, 7, false)
	}
	sort.Slice(cards, func(i, j int) bool { return cards[i].design < cards[j].design })
	_, states := marriageMinDeadwoodMemo(cards, 1)
	t.Logf("canonical memo states: %d", states)
	if states > 10000 {
		t.Fatalf("unexpected memo state count: %d", states)
	}
}

func TestMarriageSearchRejectsImpossibleTypeMultiplicity(t *testing.T) {
	cards := make([]*Card, 4)
	for i := range cards {
		cards[i] = NewCard(1, 7, false)
	}
	if got := marriageMinDeadwood(cards, 0); got != 28 {
		t.Fatalf("deadwood=%d, want 28", got)
	}
	if MarriageValidateDeclaration(cards, 0) {
		t.Fatal("impossible declaration accepted")
	}
	if MarriageHasPureSequences(cards, 0, 1) {
		t.Fatal("impossible pure sequence accepted")
	}
}

// marriageMinDeadwoodReference preserves the pre-memoization search for parity checks.
func marriageMinDeadwoodReference(cards []*Card, wildRank int) (int, int) {
	n := len(cards)
	if n == 0 {
		return 0, 0
	}
	melds := marriageGenerateMelds(cards, wildRank)
	covering := marriageCovering(n, melds)
	points := make([]int, n)
	for i, c := range cards {
		points[i] = marriageCardPoints(c, wildRank)
	}
	decided := make([]bool, n)
	iter := 0
	var dfs func() int
	dfs = func() int {
		iter++
		if iter > marriageSearchCap {
			s := 0
			for k := 0; k < n; k++ {
				if !decided[k] {
					s += points[k]
				}
			}
			return s
		}
		i := marriageFirstUndecided(decided)
		if i == -1 {
			return 0
		}
		decided[i] = true
		best := points[i] + dfs()
		decided[i] = false
		for _, mi := range covering[i] {
			m := melds[mi]
			if !marriageAllUndecided(decided, m.idx) {
				continue
			}
			marriageSetDecided(decided, m.idx, true)
			c := dfs()
			marriageSetDecided(decided, m.idx, false)
			if c < best {
				best = c
			}
		}
		return best
	}
	return dfs(), iter
}

func TestMarriageMinDeadwoodMemoizedMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(8051))
	belowCap := 0
	capReached := 0
	for sample := 0; sample < 100; sample++ {
		n := 21 + sample%2
		cards := marriageRandomPhysicalHand(rng, n)
		wildRank := 1 + rng.Intn(13)
		old, iters := marriageMinDeadwoodReference(cards, wildRank)
		got := marriageMinDeadwood(cards, wildRank)
		if iters <= marriageSearchCap {
			belowCap++
			if got != old {
				t.Fatalf("sample %d: got %d, the uncapped reference found %d (%d iterations)", sample, got, old, iters)
			}
		} else {
			capReached++
			if got > old {
				t.Fatalf("sample %d: exact result %d exceeds capped result %d", sample, got, old)
			}
		}
	}
	if belowCap == 0 {
		t.Fatal("expected at least one reference hand to finish below the search cap")
	}
	t.Logf("reference below cap: %d; cap reached: %d", belowCap, capReached)
}
