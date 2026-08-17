package games_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestMushiWildRuleReadsTheSameInBothCatalogues keeps the wild card's one
// exception saying the same thing on both surfaces.
//
// The rule -- the lightning card takes anything except a willow -- decides most
// of the mistakes in this game, and it lived only on the Web page until #5569
// put it in the CUI too. The two catalogues are separate files (`internal/i18n`
// for the CUI, `frontend/src/i18n` for the Web), so the sentence now exists
// twice with nothing tying the copies together: reword one and the other keeps
// teaching the old rule, silently and only to half the players.
func TestMushiWildRuleReadsTheSameInBothCatalogues(t *testing.T) {
	root := filepath.Join("..", "..", "..")

	read := func(path string) map[string]any {
		t.Helper()
		raw, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		return m
	}

	for _, lang := range []string{"ja", "en"} {
		cui := read("internal/i18n/locales/" + lang + "/mushi.json")
		web := read("frontend/src/i18n/locales/" + lang + "/mushi.json")

		cuiRule, ok := cui["wildRule"].(string)
		if !ok || cuiRule == "" {
			t.Fatalf("%s: internal/i18n has no mushi.wildRule", lang)
		}
		webRule, ok := web["wildRule"].(string)
		if !ok || webRule == "" {
			t.Fatalf("%s: frontend/src/i18n has no wildRule", lang)
		}
		if cuiRule != webRule {
			t.Errorf("%s: the wild rule differs between the two catalogues\n CUI: %s\n Web: %s", lang, cuiRule, webRule)
		}
	}
}
