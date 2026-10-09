//go:build test

package i18n_test

import (
	"encoding/json"
	"os"
	"testing"
)

func TestFortyThievesJapaneseCommandHelp(t *testing.T) {
	raw, err := os.ReadFile("locales/ja/fortythieves.json")
	if err != nil {
		t.Fatal(err)
	}
	var locale map[string]string
	if err := json.Unmarshal(raw, &locale); err != nil {
		t.Fatalf("decode Japanese locale: %v", err)
	}

	expected := map[string]string{
		"helpDraw":         "  d                    ストックから1枚引く",
		"helpMove":         "  m w t <col>          ウェイストからタブローへ移動",
		"helpMoveWF":       "  m w f                ウェイストから組札へ移動",
		"helpMoveTF":       "  m t <col> f          タブローから組札へ移動",
		"helpMoveTT":       "  m t <col> <idx> t <col>  タブローからタブローへ移動",
		"helpGiveUp":       "  g                    ギブアップ",
		"helpHint":         "  h                    ヒントを表示",
		"helpAutoComplete": "  ac                   オートコンプリート",
	}
	for key, want := range expected {
		got, ok := locale[key]
		if !ok {
			t.Errorf("missing %s", key)
			continue
		}
		if got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
}
