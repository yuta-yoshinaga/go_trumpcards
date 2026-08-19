package games_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestSokoWebRankingMatchesTheCuiLabels guards the Web's always-on hand-ranking
// table against the labels the showdown actually uses.
//
// The Web table is a literal list in `frontend/src/i18n/locales/<lang>/soko.json`
// (#5737); the showdown badge and the CUI both resolve `sokoHandRank<N>` from
// `internal/i18n/locales/<lang>/cui_common.json`. Both sides are internally
// consistent, so a reordered or renamed entry would teach the wrong ranking with
// nothing to notice — this compares them element by element, strongest first.
func TestSokoWebRankingMatchesTheCuiLabels(t *testing.T) {
	root := filepath.Join("..", "..", "..")

	for _, lang := range []string{"ja", "en"} {
		cui := readJSONMap(t, filepath.Join(root, "internal", "i18n", "locales", lang, "cui_common.json"))
		web := readJSONMap(t, filepath.Join(root, "frontend", "src", "i18n", "locales", lang, "soko.json"))

		ranking, ok := web["ranking"].(map[string]any)
		if !ok {
			t.Fatalf("%s/soko.json has no ranking object", lang)
		}
		rawHands, ok := ranking["hands"].([]any)
		if !ok {
			t.Fatalf("%s/soko.json ranking.hands is not a list", lang)
		}
		want := domain.SokoHandRoyalFlush + 1
		if len(rawHands) != want {
			t.Fatalf("%s lists %d hands, want %d", lang, len(rawHands), want)
		}

		// 強い順に並ぶので、i 番目は rank = RoyalFlush - i。
		for i, raw := range rawHands {
			rank := domain.SokoHandRoyalFlush - i
			key := "sokoHandRank" + strconv.Itoa(rank)
			expected, found := cui[key].(string)
			if !found || expected == "" {
				t.Fatalf("%s/cui_common.json has no %s", lang, key)
			}
			if got, _ := raw.(string); got != expected {
				t.Errorf("%s rank %d: web says %q, CUI says %q", lang, rank, got, expected)
			}
		}
	}
}

// readJSONMap reads a locale file into a generic map.
func readJSONMap(t *testing.T, path string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(path) //nolint:gosec // test-only, fixed paths
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return out
}
