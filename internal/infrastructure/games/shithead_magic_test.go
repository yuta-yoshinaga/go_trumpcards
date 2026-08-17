package games_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestShitheadMagicEffectsReadTheSameInBothCatalogues keeps the four magic-card
// effects saying the same thing on both surfaces.
//
// The Web page has had the badges and tooltips since it shipped; the CUI had no
// such wording at all until #5577, which means the sentences now exist twice in
// two separate catalogues (`internal/i18n` and `frontend/src/i18n`). Nothing
// ties the copies together, so a reworded rule would keep being taught the old
// way to whichever half nobody edited.
func TestShitheadMagicEffectsReadTheSameInBothCatalogues(t *testing.T) {
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

	// CUI key -> Web key under `magicEffect`.
	pairs := map[string]string{
		"magicEffectTwo":   "two",
		"magicEffectSeven": "seven",
		"magicEffectEight": "eight",
		"magicEffectTen":   "ten",
	}

	for _, lang := range []string{"ja", "en"} {
		cui := read("internal/i18n/locales/" + lang + "/shithead.json")
		web := read("frontend/src/i18n/locales/" + lang + "/shithead.json")

		effects, ok := web["magicEffect"].(map[string]any)
		if !ok {
			t.Fatalf("%s: frontend shithead.json has no magicEffect map", lang)
		}
		for cuiKey, webKey := range pairs {
			cuiText, ok := cui[cuiKey].(string)
			if !ok || cuiText == "" {
				t.Fatalf("%s: internal/i18n has no shithead.%s", lang, cuiKey)
			}
			webText, ok := effects[webKey].(string)
			if !ok || webText == "" {
				t.Fatalf("%s: frontend magicEffect.%s is missing", lang, webKey)
			}
			if cuiText != webText {
				t.Errorf("%s: shithead.%s differs between the catalogues\n CUI: %s\n Web: %s",
					lang, cuiKey, cuiText, webText)
			}
		}
	}
}
