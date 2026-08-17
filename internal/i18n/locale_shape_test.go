//go:build test

package i18n_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestCuiLocaleFilesAreFlatStringMaps guards the shape the loader actually
// supports.
//
// `loadTranslations` unmarshals each file into `map[string]string` and
// **skips the file on error**:
//
//	var m map[string]string
//	if err := json.Unmarshal(data, &m); err != nil {
//		continue
//	}
//
// So one nested object — the shape every frontend locale file uses — silently
// drops **that whole game's** CUI translations, and the screen fills with raw
// keys like `canasta.helpTitle`. Nothing fails; the game just stops being
// translated. This test is the missing signal (hit while adding
// `canasta.drawBlocker*`).
func TestCuiLocaleFilesAreFlatStringMaps(t *testing.T) {
	root := filepath.Join("locales")
	langs, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read locales: %v", err)
	}

	checked := 0
	for _, lang := range langs {
		if !lang.IsDir() {
			continue
		}
		dir := filepath.Join(root, lang.Name())
		files, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, f := range files {
			if f.IsDir() || filepath.Ext(f.Name()) != ".json" {
				continue
			}
			path := filepath.Join(dir, f.Name())
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			var flat map[string]string
			if err := json.Unmarshal(raw, &flat); err != nil {
				// Name the offending keys: "cannot unmarshal object" alone does not
				// say which one.
				var loose map[string]any
				if err2 := json.Unmarshal(raw, &loose); err2 == nil {
					for k, v := range loose {
						if _, ok := v.(string); !ok {
							t.Errorf("%s: key %q is not a string (%T). The CUI loader takes flat "+
								"string maps only and skips the whole file otherwise, which would "+
								"leave this game untranslated. Use dotted flat keys instead.", path, k, v)
						}
					}
					continue
				}
				t.Errorf("%s: not valid JSON: %v", path, err)
				continue
			}
			checked++
		}
	}

	// 0 件を成功と読ませない。
	if checked < 300 {
		t.Fatalf("only %d locale files checked; the walk found too few to be trusted", checked)
	}
}
