//go:build test

package domain

import (
	"os"
	"regexp"
	"sort"
	"testing"
)

// TestSutdaFrontendHandRankingMatchesDomain protects the frontend's local
// ranking list from drifting away from SutdaEvaluate's Rank definitions.
func TestSutdaFrontendHandRankingMatchesDomain(t *testing.T) {
	data, err := os.ReadFile("../../frontend/src/utils/cli/commands/sutdaCommands.ts")
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`(?s)SUTDA_HAND_RANKING_KEYS\s*= \[(.*?)\]`)
	match := re.FindSubmatch(data)
	if len(match) != 2 {
		t.Fatal("SUTDA_HAND_RANKING_KEYS not found")
	}
	keyRE := regexp.MustCompile(`'([^']+)'`)
	frontendKeys := make([]string, 0)
	for _, sub := range keyRE.FindAllSubmatch(match[1], -1) {
		frontendKeys = append(frontendKeys, string(sub[1]))
	}

	type ranked struct {
		name string
		rank int
	}
	all := make([]ranked, 0)
	for a := 1; a <= SutdaMonthCnt; a++ {
		for b := a; b <= SutdaMonthCnt; b++ {
			for ca := 1; ca <= SutdaCopiesPerMonth; ca++ {
				for cb := 1; cb <= SutdaCopiesPerMonth; cb++ {
					hand := SutdaEvaluate(NewCard(a, ca, false), NewCard(b, cb, false))
					all = append(all, ranked{name: hand.Name, rank: hand.Rank})
				}
			}
		}
	}
	best := map[string]int{}
	for _, item := range all {
		if item.rank > best[item.name] {
			best[item.name] = item.rank
		}
	}
	want := make([]ranked, 0, len(best))
	for name, rank := range best {
		if name != "none" {
			want = append(want, ranked{name: name, rank: rank})
		}
	}
	sort.Slice(want, func(i, j int) bool { return want[i].rank > want[j].rank })
	wantKeys := make([]string, len(want))
	for i, item := range want {
		wantKeys[i] = item.name
	}
	if len(frontendKeys) != len(wantKeys) {
		t.Fatalf("frontend has %d ranking entries, domain has %d: %v", len(frontendKeys), len(wantKeys), wantKeys)
	}
	for i := range wantKeys {
		if frontendKeys[i] != wantKeys[i] {
			t.Fatalf("ranking %d: frontend=%s domain=%s", i+1, frontendKeys[i], wantKeys[i])
		}
	}
}
