package games_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestKlondikeVegasFormulaMatchesTheLocales guards the on-screen explanation of
// the Vegas score against the domain constants that actually compute it.
//
// `GetScore` returns `KlondikeVegasBuyIn + KlondikeVegasPerCard * cards`, and
// the Klondike page prints that formula so a player choosing "Vegas" can tell
// why the score starts negative (#5493). **The text and the arithmetic live in
// different languages in different directories**, so nothing stops one from
// being edited alone -- and a scoring rule that disagrees with the number on
// screen is worse than no explanation at all.
//
// The page interpolates the numbers from `KlondikeVegas` in
// `frontend/src/types/phases.ts`, so this checks that file too: the locale
// strings only carry the placeholders.
func TestKlondikeVegasFormulaMatchesTheLocales(t *testing.T) {
	root := filepath.Join("..", "..", "..")

	t.Run("the TypeScript constants match the domain", func(t *testing.T) {
		src, err := os.ReadFile(filepath.Join(root, "frontend", "src", "types", "phases.ts"))
		if err != nil {
			t.Fatalf("read phases.ts: %v", err)
		}
		text := string(src)
		for _, want := range []struct {
			field string
			value int
		}{
			{"BUY_IN", domain.KlondikeVegasBuyIn},
			{"PER_CARD", domain.KlondikeVegasPerCard},
		} {
			line := want.field + ": " + strconv.Itoa(want.value) + ","
			if !strings.Contains(text, line) {
				t.Errorf("frontend/src/types/phases.ts should carry `%s` (KlondikeVegas.%s must equal the domain constant)",
					line, want.field)
			}
		}
	})

	// 文言側は数値を書かず、プレースホルダだけを持つこと。数値を直接書くと、
	// 定数を直しても文言が古いまま残る。
	t.Run("the locale strings interpolate rather than hardcode", func(t *testing.T) {
		for _, loc := range []string{"ja", "en"} {
			path := filepath.Join(root, "frontend", "src", "i18n", "locales", loc, "klondike.json")
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			var d map[string]any
			if err := json.Unmarshal(raw, &d); err != nil {
				t.Fatalf("parse %s: %v", path, err)
			}
			text, ok := d["vegasFormula"].(string)
			if !ok {
				t.Fatalf("%s has no vegasFormula string", path)
			}
			for _, ph := range []string{"{{buyIn}}", "{{perCard}}"} {
				if !strings.Contains(text, ph) {
					t.Errorf("%s vegasFormula must interpolate %s, got %q", loc, ph, text)
				}
			}
			if strings.Contains(text, strconv.Itoa(-domain.KlondikeVegasBuyIn)) {
				t.Errorf("%s vegasFormula hardcodes the buy-in; interpolate it instead: %q", loc, text)
			}
		}
	})
}
