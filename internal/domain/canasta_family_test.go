//go:build test && (!js || !wasm || extra7)

package domain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCanastaFamilyCardValue_GoldenVectors(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "frontend", "src", "utils", "__fixtures__", "canastaFamilyCardValue.golden.json"))
	if err != nil {
		t.Fatalf("read golden vectors: %v", err)
	}
	var golden struct {
		Cases []struct {
			Name, Design  string
			Value, Points int
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("parse golden vectors: %v", err)
	}
	if len(golden.Cases) != 53 {
		t.Fatalf("got %d vectors, want 53", len(golden.Cases))
	}
	designs := map[string]int{"JOKER": CardDesignJoker, "SPADE": CardDesignSpade, "CLOVER": CardDesignClover, "HEART": CardDesignHeart, "DIAMOND": CardDesignDiamond}
	for _, c := range golden.Cases {
		design, ok := designs[c.Design]
		if !ok {
			t.Fatalf("%s: unknown design %q", c.Name, c.Design)
		}
		if got := CanastaFamilyCardValue(NewCard(design, c.Value, true)); got != c.Points {
			t.Errorf("%s: got %d, want %d", c.Name, got, c.Points)
		}
	}
}
