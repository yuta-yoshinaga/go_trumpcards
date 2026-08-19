package games_test

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestMinibridgeTeamTagsAgreeAcrossTheUIs guards the two team labels.
//
// The CUI has always printed `team` on every seat line; the Web page now does
// too (#5761). Both read the same `team` value, so a swapped ally test would
// look self-consistent on each screen and only disagree with the other.
func TestMinibridgeTeamTagsAgreeAcrossTheUIs(t *testing.T) {
	root := filepath.Join("..", "..", "..")

	cui := readFileForTest(t, filepath.Join(root, "internal", "adapter", "presenter", "MinibridgeCuiPresenter.go"))
	if !strings.Contains(cui, `"team", strconv.Itoa(player.GetTeam())`) {
		t.Fatal("the CUI no longer prints the seat's team, so the Web tag has nothing to agree with")
	}

	web := readFileForTest(t, filepath.Join(root, "frontend", "src", "pages", "MinibridgePage.tsx"))
	if !regexp.MustCompile(`t\('header\.teamTag', \{ team: String\(p\.team\) \}\)`).MatchString(web) {
		t.Error("MinibridgePage.tsx no longer labels each seat from p.team")
	}
	// **味方判定は人間のチームとの一致。**逆にすると味方と相手が入れ替わる。
	if !regexp.MustCompile(`p\.team === humanTeam \? t\('header\.teamAllyAria'\)`).MatchString(web) {
		t.Error("the ally reading is no longer keyed on matching the human's team")
	}
}
