//go:build test

package domain

import (
	"math/rand"
	"sort"
	"testing"
)

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
	for i := 0; i < 80; i++ {
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

func TestMarriageMinDeadwoodCanonicalMatchesMemoReference(t *testing.T) {
	rng := rand.New(rand.NewSource(805122))
	for sample := 0; sample < 320; sample++ {
		n := 21 + sample%2
		cards := make([]*Card, n)
		wildRank := 0
		if sample%10 == 0 {
			wildRank = 1 + rng.Intn(13)
		}
		for i := range cards {
			// Include frequent broad hands and a smaller duplicate/wild stress set.
			design, value := 1+rng.Intn(4), 1+rng.Intn(13)
			if rng.Intn(24) == 0 || sample%10 == 0 && rng.Intn(3) == 0 {
				design, value = CardDesignJoker, CardValueJoker
			}
			cards[i] = NewCard(design, value, false)
		}
		want := marriageMinDeadwoodMemoReference(cards, wildRank)
		if got := marriageMinDeadwood(cards, wildRank); got != want {
			t.Fatalf("sample %d: got %d, reference %d", sample, got, want)
		}
	}
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
		cards := make([]*Card, n)
		for i := range cards {
			design := 1 + rng.Intn(4)
			value := 1 + rng.Intn(13)
			if rng.Intn(24) == 0 {
				design, value = CardDesignJoker, CardValueJoker
			}
			cards[i] = NewCard(design, value, false)
		}
		wildRank := 1 + rng.Intn(13)
		old, iters := marriageMinDeadwoodReference(cards, wildRank)
		got := marriageMinDeadwood(cards, wildRank)
		if iters <= marriageSearchCap {
			belowCap++
			if got != old {
				t.Fatalf("sample %d: got %d, reference %d (%d iterations)", sample, got, old, iters)
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
