package games_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestGinRummyCpuPolicyConstantsMatchTheFrontend guards the difficulty
// descriptions in the settings panel against the code that acts on them.
//
// `cpuDiscardOrKnock` / `cpuDraw` branch on difficulty: Easy knocks at the legal
// limit and takes the discard at random, Normal wants a 7 and a real gain, Hard
// waits for a 5 and takes any gain. None of that was visible to the player, and
// the panel now quotes the numbers (#5500). **They are written twice, in two
// languages**, so this pins the TypeScript copy to the domain.
func TestGinRummyCpuPolicyConstantsMatchTheFrontend(t *testing.T) {
	path := filepath.Join("..", "..", "..", "frontend", "src", "types", "phases.ts")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read phases.ts: %v", err)
	}
	text := string(src)

	for _, want := range []struct {
		field string
		value int
	}{
		{"KNOCK_DEADWOOD_NORMAL", domain.GinRummyKnockDeadwoodNormal},
		{"KNOCK_DEADWOOD_HARD", domain.GinRummyKnockDeadwoodHard},
		{"DISCARD_GAIN_NORMAL", domain.GinRummyDiscardGainNormal},
		{"EASY_PICK_ONE_IN", domain.GinRummyEasyPickOneIn},
		{"KNOCK_THRESHOLD", domain.GinRummyKnockThreshold},
	} {
		line := want.field + ": " + strconv.Itoa(want.value) + ","
		if !strings.Contains(text, line) {
			t.Errorf("frontend/src/types/phases.ts should carry `%s` in GinRummyCpu (domain says %d)",
				line, want.value)
		}
	}

	// **順序も説明の一部。** Hard のほうが Normal より辛抱強い、という説明が
	// 逆転していないこと。
	if domain.GinRummyKnockDeadwoodHard >= domain.GinRummyKnockDeadwoodNormal {
		t.Errorf("Hard should knock at a stricter deadwood than Normal (hard=%d normal=%d)",
			domain.GinRummyKnockDeadwoodHard, domain.GinRummyKnockDeadwoodNormal)
	}
	if domain.GinRummyKnockDeadwoodNormal >= domain.GinRummyKnockThreshold {
		t.Errorf("Normal should knock below the legal limit (normal=%d limit=%d)",
			domain.GinRummyKnockDeadwoodNormal, domain.GinRummyKnockThreshold)
	}
}
