//go:build test

package domain

import (
	"os"
	"regexp"
	"strconv"
	"testing"
)

// TestCostlyColoursFrontendJackOrDeuceMatchesDomain protects the frontend's
// two-value badge condition from drifting away from the domain predicate.
func TestCostlyColoursFrontendJackOrDeuceMatchesDomain(t *testing.T) {
	data, err := os.ReadFile("../../frontend/src/pages/CostlyColoursPage.tsx")
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`card\.value === (\d+)\s*\|\|\s*card\.value === (\d+)`)
	match := re.FindSubmatch(data)
	if len(match) != 3 {
		t.Fatal("Costly Colours J / 2 condition not found")
	}
	frontendValues := map[int]bool{}
	for _, raw := range match[1:] {
		value, err := strconv.Atoi(string(raw))
		if err != nil {
			t.Fatalf("frontend special value is not an integer: %v", err)
		}
		frontendValues[value] = true
	}
	domainValues := map[int]bool{}
	for value := 1; value <= 13; value++ {
		if CostlyIsJackOrDeuce(NewCard(CardDesignSpade, value, false)) {
			domainValues[value] = true
		}
	}
	if len(frontendValues) != len(domainValues) {
		t.Fatalf("frontend values=%v, domain values=%v", frontendValues, domainValues)
	}
	for value := range domainValues {
		if !frontendValues[value] {
			t.Fatalf("frontend is missing domain special value %d", value)
		}
	}
}

// TestCostlyColoursFrontendJackDeucePointsMatchDomain protects the frontend's
// trump and plain Jack / deuce badge points from drifting away from the domain.
func TestCostlyColoursFrontendJackDeucePointsMatchDomain(t *testing.T) {
	data, err := os.ReadFile("../../frontend/src/pages/CostlyColoursPage.tsx")
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`state\.turnUp\?\.design === card\.design \? (\d+) : (\d+)`)
	match := re.FindSubmatch(data)
	if len(match) != 3 {
		t.Fatal("Costly Colours J / 2 points condition not found")
	}
	trumpPoints, err := strconv.Atoi(string(match[1]))
	if err != nil {
		t.Fatalf("frontend trump points are not an integer: %v", err)
	}
	plainPoints, err := strconv.Atoi(string(match[2]))
	if err != nil {
		t.Fatalf("frontend plain points are not an integer: %v", err)
	}
	if trumpPoints != CostlyTrumpJackDeucePoints {
		t.Fatalf("frontend trump points=%d, domain points=%d", trumpPoints, CostlyTrumpJackDeucePoints)
	}
	if plainPoints != CostlyPlainJackDeucePoints {
		t.Fatalf("frontend plain points=%d, domain points=%d", plainPoints, CostlyPlainJackDeucePoints)
	}
}
