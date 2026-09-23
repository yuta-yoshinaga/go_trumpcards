//go:build test

package domain

import (
	"os"
	"regexp"
	"strconv"
	"testing"
)

// TestCometFrontendWildCardMatchesDomain protects the frontend copy of the
// Comet wild-card constants from drifting away from the domain constants.
func TestCometFrontendWildCardMatchesDomain(t *testing.T) {
	data, err := os.ReadFile("../../frontend/src/pages/CometPage.tsx")
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`isCometWild\s*=.*?card\.design === '([^']+)'\s*&&\s*card\.value === (\d+)`)
	match := re.FindSubmatch(data)
	if len(match) != 3 {
		t.Fatal("isCometWild constants not found")
	}
	value, err := strconv.Atoi(string(match[2]))
	if err != nil {
		t.Fatalf("frontend wild value is not an integer: %v", err)
	}
	designs := []string{"SPADE", "CLOVER", "HEART", "DIAMOND"}
	if CometWildDesign < 1 || CometWildDesign > len(designs) {
		t.Fatalf("CometWildDesign is outside standard suit range: %d", CometWildDesign)
	}
	if string(match[1]) != designs[CometWildDesign-1] {
		t.Fatalf("frontend wild design=%s, domain design=%s", match[1], designs[CometWildDesign-1])
	}
	if value != CometWildValue {
		t.Fatalf("frontend wild value=%d, domain value=%d", value, CometWildValue)
	}
}
