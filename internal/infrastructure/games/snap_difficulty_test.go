package games_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestSnapDifficultyHintsMatchTheReactionTimes guards the numbers the settings
// panel quotes.
//
// The panel now tells the player roughly how fast each difficulty reacts
// (#5763). Those seconds come from `drawReactionMs`'s means, which live only in
// the domain — a copy in the locale file would keep promising 1.4s long after
// the distribution moved, and nothing else would notice.
func TestSnapDifficultyHintsMatchTheReactionTimes(t *testing.T) {
	root := filepath.Join("..", "..", "..")

	src := readFileForTest(t, filepath.Join(root, "internal", "domain", "Snap.go"))
	means := map[string]float64{}
	for _, m := range regexp.MustCompile(`mean, sd = (\d+)\.0, \d+\.0`).FindAllStringSubmatch(src, -1) {
		v, err := strconv.Atoi(m[1])
		if err != nil {
			t.Fatalf("parse a mean: %v", err)
		}
		means[m[1]] = float64(v) / 1000
	}
	// easy と hard が switch に、normal が既定値として書かれている。
	if len(means) != 2 {
		t.Fatalf("found %d means in drawReactionMs, want the easy and hard cases", len(means))
	}

	raw, err := os.ReadFile(filepath.Join(root, "frontend", "src", "i18n", "locales", "ja", "snap.json")) //nolint:gosec // test-only
	if err != nil {
		t.Fatalf("read the locale: %v", err)
	}
	var locale struct {
		Actions map[string]string `json:"actions"`
	}
	if err := json.Unmarshal(raw, &locale); err != nil {
		t.Fatalf("parse the locale: %v", err)
	}

	// 1400 → "1.4秒", 500 → "0.5秒" のように、実際の平均が文言に現れること。
	for msKey, seconds := range means {
		want := strconv.FormatFloat(seconds, 'f', 1, 64) + "秒"
		hint := locale.Actions["easyHint"]
		if msKey == "500" {
			hint = locale.Actions["hardHint"]
		}
		if !strings.Contains(hint, want) {
			t.Errorf("the %sms difficulty is described as %q, which does not mention %s", msKey, hint, want)
		}
		if !strings.Contains(locale.Actions["difficultyTip"], want) {
			t.Errorf("the tooltip does not mention %s for the %sms difficulty", want, msKey)
		}
	}
}
