//go:build test

package domain

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
)

// TestBalootCardPointsFrontendGuard compares the domain table with the two
// tables used by the frontend utility for every card in both modes.
func TestBalootCardPointsFrontendGuard(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "frontend", "src", "utils", "balootPoints.ts"))
	if err != nil {
		t.Fatal(err)
	}

	if !regexp.MustCompile(`mode\s*===\s*2`).Match(source) {
		t.Fatal("frontend Baloot trump check must use mode === 2")
	}

	mapBodies := regexp.MustCompile(`\(\{([^{}]*)\}\s+as\s+Record<number,\s*number>\)\[card\.value\]\s*\?\?\s*0`).FindAllSubmatch(source, -1)
	if len(mapBodies) != 2 {
		t.Fatalf("found %d Baloot frontend point maps, want 2", len(mapBodies))
	}
	parseMap := func(body []byte) map[int]int {
		points := make(map[int]int)
		entries := regexp.MustCompile(`(\d+)\s*:\s*(\d+)`).FindAllSubmatch(body, -1)
		for _, entry := range entries {
			value, valueErr := strconv.Atoi(string(entry[1]))
			point, pointErr := strconv.Atoi(string(entry[2]))
			if valueErr != nil || pointErr != nil {
				t.Fatalf("invalid frontend Baloot point entry %q", entry)
			}
			points[value] = point
		}
		return points
	}
	trumpPoints := parseMap(mapBodies[0][1])
	nonTrumpPoints := parseMap(mapBodies[1][1])

	designs := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	values := []int{1, 7, 8, 9, 10, 11, 12, 13}
	for _, design := range designs {
		for _, value := range values {
			card := NewCard(design, value, false)
			wantSun := nonTrumpPoints[value]
			if got := BalootCardPoints(card, BalootModeSun, 0); got != wantSun {
				t.Errorf("Sun %d/%d: got %d, want %d", design, value, got, wantSun)
			}
			if got := BalootCardPoints(card, BalootModeHokom, design); got != trumpPoints[value] {
				t.Errorf("Hokom trump %d/%d: got %d, want %d", design, value, got, trumpPoints[value])
			}
			if got := BalootCardPoints(card, BalootModeHokom, (design%len(designs))+1); got != nonTrumpPoints[value] {
				t.Errorf("Hokom non-trump %d/%d: got %d, want %d", design, value, got, nonTrumpPoints[value])
			}
		}
	}
}
