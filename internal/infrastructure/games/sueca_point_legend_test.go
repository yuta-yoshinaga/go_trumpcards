package games_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestSuecaWebPointLegendMatchesTheDomain guards the Web card-point legend
// against the values the domain actually scores.
//
// suecaCardPoints (internal/domain/Sueca.go) is unexported, so the legend
// (`frontend/src/i18n/locales/*/sueca.json`, `pointLegend.*`) spells the same
// numbers as **plain text** (#5642). Change the scoring and the two disagree
// with nothing to notice -- the same drift #5497 found in the Baccarat panel.
//
// Every value 1..13 is walked, so a newly-scoring card cannot slip into the
// "others = 0" row unnoticed, and the round total is summed rather than copied.
func TestSuecaWebPointLegendMatchesTheDomain(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	// value -> the legend key that must state its points.
	keyOf := map[int]string{1: "aceValue", 7: "sevenValue", 13: "kingValue", 11: "jackValue", 12: "queenValue"}

	want := map[string]int{}
	total := 0
	const suitCnt = 4
	for value := 1; value <= 13; value++ {
		pts := domain.SuecaCardPoints(value)
		total += pts * suitCnt
		if key, ok := keyOf[value]; ok {
			want[key] = pts
			continue
		}
		if pts != 0 {
			t.Fatalf("value %d scores %d but the legend lumps it under 'others' (0)", value, pts)
		}
	}
	want["othersValue"] = 0

	for _, loc := range []string{"ja", "en"} {
		path := filepath.Join(root, "frontend", "src", "i18n", "locales", loc, "sueca.json")
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		var d map[string]any
		if err := json.Unmarshal(raw, &d); err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		legend, ok := d["pointLegend"].(map[string]any)
		if !ok {
			t.Fatalf("%s has no pointLegend object", path)
		}
		for key, pts := range want {
			line, ok := legend[key].(string)
			if !ok {
				t.Errorf("%s pointLegend.%s missing", loc, key)
				continue
			}
			// **部分一致では通っても何も保証しない** (#6009)。0 は "10点" の
			// 0 でも通ってしまうので、桁の境界まで見る。
			if !statesNumber(line, pts) {
				t.Errorf("%s pointLegend.%s = %q, want it to state %d", loc, key, line, pts)
			}
		}
		note, ok := legend["note"].(string)
		if !ok {
			t.Fatalf("%s pointLegend.note missing", loc)
		}
		// 合計 120 点と勝利ライン 61 点も本文に書いてある。数字だけずれても気づけない。
		for _, n := range []int{total, domain.SuecaWinPoints} {
			if !statesNumber(note, n) {
				t.Errorf("%s pointLegend.note = %q, want it to state %d", loc, note, n)
			}
		}
	}
}
