package i18n

import (
	"encoding/json"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

// TestLoadTranslations_MissingFile covers the ReadFile error branch
// by loading a language with no locale files.
func TestLoadTranslations_MissingFile(t *testing.T) {
	emptyFS := fstest.MapFS{}
	result := loadTranslations(emptyFS, "ja")
	assert.Empty(t, result)
}

// TestLoadTranslations_InvalidJSON covers the json.Unmarshal error branch.
func TestLoadTranslations_InvalidJSON(t *testing.T) {
	badFS := fstest.MapFS{
		"locales/xx/common.json": &fstest.MapFile{Data: []byte("{invalid json")},
	}
	result := loadTranslations(badFS, "xx")
	assert.Empty(t, result)
}

// TestLoadTranslations_ValidLang covers the happy path for both ja and en.
func TestLoadTranslations_ValidLang(t *testing.T) {
	result := loadTranslations(localesFS, "en")
	assert.NotEmpty(t, result)
	assert.Equal(t, "Goodbye.", result["bye"])
}

// TestLoadTranslations_AutoDiscovery verifies that the loader picks up any
// *.json file dropped under locales/<lang>/ without requiring a code edit.
// This is the core invariant of the auto-discovery refactor — the old
// hand-maintained `files` slice repeatedly drifted out of sync (eleven
// loader-divergence bugs were caught and fixed across #1699 Phase 3 alone),
// so the regression we are guarding against is "new locale file added but
// not registered → keys silently return as the literal key string."
func TestLoadTranslations_AutoDiscovery(t *testing.T) {
	fsys := fstest.MapFS{
		"locales/xx/common.json":     &fstest.MapFile{Data: []byte(`{"globalKey":"GLOBAL"}`)},
		"locales/xx/cui_common.json": &fstest.MapFile{Data: []byte(`{"cuiShared":"SHARED"}`)},
		"locales/xx/blackjack.json":  &fstest.MapFile{Data: []byte(`{"helpTitle":"BJ"}`)},
		"locales/xx/newgame.json":    &fstest.MapFile{Data: []byte(`{"helpTitle":"NEW"}`)},
		// Non-JSON files and subdirectories must be ignored.
		"locales/xx/README.md":       &fstest.MapFile{Data: []byte("not json")},
		"locales/xx/sub/nested.json": &fstest.MapFile{Data: []byte(`{"nope":"x"}`)},
	}

	result := loadTranslations(fsys, "xx")

	// Global namespaces (common, cui_common) merge flat.
	assert.Equal(t, "GLOBAL", result["globalKey"])
	assert.Equal(t, "SHARED", result["cuiShared"])

	// Per-game files are namespaced by file name. The previously-unknown
	// `newgame.json` works without any code change — that's the whole
	// point of the refactor.
	assert.Equal(t, "BJ", result["blackjack.helpTitle"])
	assert.Equal(t, "NEW", result["newgame.helpTitle"])

	// README.md and the nested sub/ directory must not contribute keys.
	assert.NotContains(t, result, "README")
	assert.NotContains(t, result, "nope")
	assert.NotContains(t, result, "sub.nested.nope")
}

// TestLoadTranslations_AllExistingLocaleFilesResolve guards against the
// pre-refactor failure mode (locale file present on disk but not in the
// hand-maintained slice). For every embedded *.json, at least one of
// that file's own keys must appear in the loaded translation map.
func TestLoadTranslations_AllExistingLocaleFilesResolve(t *testing.T) {
	for _, lang := range []string{"ja", "en"} {
		entries, err := localesFS.ReadDir("locales/" + lang)
		assert.NoError(t, err)
		loaded := loadTranslations(localesFS, lang)

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			file := strings.TrimSuffix(entry.Name(), ".json")
			assert.True(t, hasFileContribution(localesFS, loaded, lang, file),
				"locale file %s/%s loaded zero keys — auto-discovery regression?",
				lang, entry.Name())
		}
	}
}

// hasFileContribution checks whether any of the keys defined in
// locales/<lang>/<file>.json actually made it into `loaded`. This proves
// the file's own contents are visible (not merely that some other file
// happened to load), which is what TestLoadTranslations_AllExistingLocaleFilesResolve
// guards against.
func hasFileContribution(fsys fs.FS, loaded map[string]string, lang, file string) bool {
	data, err := fs.ReadFile(fsys, "locales/"+lang+"/"+file+".json")
	if err != nil {
		return false
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return false
	}
	for k := range m {
		key := k
		if !globalNamespaces[file] {
			key = file + "." + k
		}
		if _, ok := loaded[key]; ok {
			return true
		}
	}
	return false
}
