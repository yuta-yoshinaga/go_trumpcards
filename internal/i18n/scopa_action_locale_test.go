//go:build test

package i18n_test

import (
	"encoding/json"
	"os"
	"testing"
)

func TestScopaJapaneseCaptureActionUsesNaturalWordOrder(t *testing.T) {
	data, err := os.ReadFile("locales/ja/scopa.json")
	if err != nil {
		t.Fatal(err)
	}
	var locale map[string]string
	if err := json.Unmarshal(data, &locale); err != nil {
		t.Fatal(err)
	}
	if got, want := locale["actionCapture"], "{{played}}を出して{{count}}枚を捕獲{{suffix}}"; got != want {
		t.Fatalf("actionCapture = %q, want %q", got, want)
	}
	if got, want := locale["actionScopaSuffix"], "（スコパ！）"; got != want {
		t.Fatalf("actionScopaSuffix = %q, want %q", got, want)
	}
}
