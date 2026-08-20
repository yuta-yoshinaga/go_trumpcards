package games_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestSlapjackDifficultyLabelsMatchTheWebSelect keeps the CUI's difficulty
// wording identical to the web's.
//
// #5579 asks for the CUI to show the same value the web select already shows.
// Same value is not the same as same wording: the two catalogues are separate
// files, so "Normal" on one screen and "ふつう" on the other would both look
// correct in isolation while describing one setting two ways.
func TestSlapjackDifficultyLabelsMatchTheWebSelect(t *testing.T) {
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

	pairs := map[string]string{
		"difficultyEasy":   "easy",
		"difficultyNormal": "normal",
		"difficultyHard":   "hard",
	}

	for _, lang := range []string{"ja", "en"} {
		cui := read("internal/i18n/locales/" + lang + "/slapjack.json")
		web := read("frontend/src/i18n/locales/" + lang + "/slapjack.json")

		settings, ok := web["settings"].(map[string]any)
		if !ok {
			t.Fatalf("%s: frontend slapjack.json has no settings map", lang)
		}
		labels, ok := settings["difficulty"].(map[string]any)
		if !ok {
			t.Fatalf("%s: frontend settings has no difficulty map", lang)
		}
		for cuiKey, webKey := range pairs {
			cuiText, ok := cui[cuiKey].(string)
			if !ok || cuiText == "" {
				t.Fatalf("%s: internal/i18n has no slapjack.%s", lang, cuiKey)
			}
			webText, ok := labels[webKey].(string)
			if !ok || webText == "" {
				t.Fatalf("%s: frontend settings.difficulty.%s is missing", lang, webKey)
			}
			if cuiText != webText {
				t.Errorf("%s: slapjack.%s differs from the web select\n CUI: %s\n Web: %s",
					lang, cuiKey, cuiText, webText)
			}
		}
	}
}
