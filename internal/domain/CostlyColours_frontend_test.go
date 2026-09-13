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
