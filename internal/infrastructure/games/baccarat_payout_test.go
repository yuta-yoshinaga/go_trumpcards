package games_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestBaccaratWebPayoutRefMatchesTheDomain guards the Web payout panel against
// the constants that actually pay out.
//
// The CUI table interpolates BaccaratTiePayoutRate / BacPairPayoutRate /
// BaccaratCommissionRate, so it cannot drift. The Web panel
// (`frontend/src/i18n/locales/*/baccarat.json`, `payoutRef.*`) spells the same
// numbers as **plain text**, which is what the CUI table was written to match
// (#5497). Change a rate and the two disagree, with nothing to notice.
func TestBaccaratWebPayoutRefMatchesTheDomain(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	// tie 8 -> "(8:1)", pair 11 -> "(11:1)", banker keeps 5% commission -> "0.95:1".
	want := map[string]string{
		"tie":        "(" + strconv.Itoa(domain.BaccaratTiePayoutRate) + ":1)",
		"playerPair": "(" + strconv.Itoa(domain.BacPairPayoutRate) + ":1)",
		"bankerPair": "(" + strconv.Itoa(domain.BacPairPayoutRate) + ":1)",
		"bankerWin":  "(" + strconv.FormatFloat(1-float64(domain.BaccaratCommissionRate)/100, 'g', -1, 64) + ":1)",
		"playerWin":  "(1:1)",
	}

	for _, loc := range []string{"ja", "en"} {
		path := filepath.Join(root, "frontend", "src", "i18n", "locales", loc, "baccarat.json")
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		var d map[string]any
		if err := json.Unmarshal(raw, &d); err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		ref, ok := d["payoutRef"].(map[string]any)
		if !ok {
			t.Fatalf("%s has no payoutRef object", path)
		}
		for key, fragment := range want {
			line, ok := ref[key].(string)
			if !ok {
				t.Errorf("%s payoutRef.%s missing", loc, key)
				continue
			}
			// **数字の境界まで見る** (#6009)。括弧付きなので今の文言では
			// 部分一致の穴は無いが、"(18:1)" のような文言に変わった日に
			// "(8:1)" が通ってしまう形は残さない。
			if !statesText(line, fragment) {
				t.Errorf("%s payoutRef.%s should state %s (from the domain constants), got %q",
					loc, key, fragment, line)
			}
		}
	}
}
