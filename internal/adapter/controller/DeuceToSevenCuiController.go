//go:build !js || !wasm || casino

package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DeuceToSevenCuiController wires the 2-7 Triple Draw interactor to the CUI
// command loop.
type DeuceToSevenCuiController struct {
	di usecase.DeuceToSevenInteractorIF
}

// NewDeuceToSevenCuiController constructs the controller.
func NewDeuceToSevenCuiController(di usecase.DeuceToSevenInteractorIF) *DeuceToSevenCuiController {
	return &DeuceToSevenCuiController{di: di}
}

// Exec runs a single CUI command. Supported commands:
//
//	r / reset        — start a new hand
//	e / exchange     — draw with the given 0-indexed card positions
//	s / stand        — pat (draw 0 cards)
//	b / bet <n>      — bet n chips
//	c / call         — call
//	ra / raise <n>   — raise by n chips
//	f / fold         — fold
//	ck / check       — check
//	a / allin        — all-in
//	bl / bettinglimit <n>    — 0 Fixed, 1 PotLimit, 2 NoLimit
//	scc / setcpucount <n>    — CPU opponent count (1..3)
//	mai / metaai <0|1>       — toggle meta-AI
//	log / l                  — action log
func (dcc *DeuceToSevenCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return dcc.di.Reset() },
		[]string{
			"e", "exchange", "s", "stand", "b", "bet", "c", "call", "ra", "raise",
			"f", "fold", "ck", "check", "a", "allin",
			"h", "hint",
			"bl", "bettinglimit", "scc", "setcpucount", "mai", "metaai",
			"log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "e", "exchange":
				indices, skipped := cuiutil.ParseBoundedIntSlice(args, 0, domain.DeuceToSevenHandSize-1)
				// Refuse before playing. PrependSkippedWarning ran the move first and
				// put the warning above the new board, so a mistyped index was dropped
				// and the remaining ones played as a different, legal move (issue #5390).
				if len(skipped) > 0 {
					return invalidArg("invalidCardIndex", "val", strings.Join(skipped, ", ")), true
				}
				return dcc.di.Exchange(indices), true
			case "s", "stand":
				return dcc.di.Stand(), true
			case "h", "hint":
				return dcc.di.Hint(), true
			case "b", "bet":
				amount := parseCuiAmount(args)
				return dcc.di.Action(domain.DeuceToSevenActionBet, amount, 0), true
			case "c", "call":
				return dcc.di.Action(domain.DeuceToSevenActionCall, 0, 0), true
			case "ra", "raise":
				amount := parseCuiAmount(args)
				return dcc.di.Action(domain.DeuceToSevenActionRaise, amount, 0), true
			case "f", "fold":
				return dcc.di.Action(domain.DeuceToSevenActionFold, 0, 0), true
			case "ck", "check":
				return dcc.di.Action(domain.DeuceToSevenActionCheck, 0, 0), true
			case "a", "allin":
				return dcc.di.Action(domain.DeuceToSevenActionAllIn, 0, 0), true
			case "bl", "bettinglimit":
				return cuiutil.WithParsedIntKeys(args, "bettingLimitTypeRequired0Fixed1Potlimit2Nolimit", "invalidBettingLimit02", 0, 2, func(v int) string {
					cfg := dcc.di.GetConfig()
					cfg.BettingLimit = domain.BettingLimitType(v)
					return dcc.di.ResetWithConfig(cfg, nil)
				})
			case "scc", "setcpucount":
				return cuiutil.WithParsedIntKeys(args, "cpuPlayerCountRequired", "invalidCpuPlayerCount13", domain.DeuceToSevenCpuCountMin, domain.DeuceToSevenCpuCountMax, func(v int) string {
					cfg := dcc.di.GetConfig()
					cfg.CpuCount = v
					return dcc.di.ResetWithConfig(cfg, nil)
				})
			case "mai", "metaai":
				if len(args) < 1 {
					return i18n.T("metaAIRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 1 {
					return invalidArg("invalidMetaAI", "val", args[0]), true
				}
				cfg := dcc.di.GetConfig()
				cfg.CpuMetaAI = v == 1
				return dcc.di.ResetWithConfig(cfg, nil), true
			default:
				return handleCuiLog(cmd, dcc.di.ActionLog)
			}
		},
	)
}
