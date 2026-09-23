//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func trappolaRoundBreakdownParams(thirds [domain.TrappolaTeamCnt]int, leadPlayerIdx int) map[string]string {
	lastTeam := domain.TrappolaTeamOf(leadPlayerIdx)
	return map[string]string{
		"a":        trappolaTeamLabel(0),
		"athird":   strconv.Itoa(thirds[0]),
		"b":        trappolaTeamLabel(1),
		"bthird":   strconv.Itoa(thirds[1]),
		"lastteam": trappolaTeamLabel(lastTeam),
	}
}
