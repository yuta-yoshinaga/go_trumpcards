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

// BadugiCuiController wires the Badugi interactor to the CUI command loop.
type BadugiCuiController struct {
	bi usecase.BadugiInteractorIF
}

// NewBadugiCuiController constructs the controller.
func NewBadugiCuiController(bi usecase.BadugiInteractorIF) *BadugiCuiController {
	return &BadugiCuiController{bi: bi}
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
func (bcc *BadugiCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return bcc.bi.Reset() },
		[]string{
			"e", "exchange", "s", "stand", "b", "bet", "c", "call", "ra", "raise",
			"f", "fold", "ck", "check", "a", "allin",
			"bl", "bettinglimit", "scc", "setcpucount", "mai", "metaai",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "e", "exchange":
				indices, skipped := cuiutil.ParseBoundedIntSlice(args, 0, domain.BadugiHandSize-1)
				// Refuse before playing. PrependSkippedWarning ran the move first and
				// put the warning above the new board, so a mistyped index was dropped
				// and the remaining ones played as a different, legal move (issue #5390).
				if len(skipped) > 0 {
					return invalidArg("invalidCardIndex", "val", strings.Join(skipped, ", ")), true
				}
				return bcc.bi.Exchange(indices), true
			case "s", "stand":
				return bcc.bi.Stand(), true
			case "b", "bet":
				amount := parseCuiAmount(args)
				return bcc.bi.Action(domain.BadugiActionBet, amount, 0), true
			case "c", "call":
				return bcc.bi.Action(domain.BadugiActionCall, 0, 0), true
			case "ra", "raise":
				amount := parseCuiAmount(args)
				return bcc.bi.Action(domain.BadugiActionRaise, amount, 0), true
			case "f", "fold":
				return bcc.bi.Action(domain.BadugiActionFold, 0, 0), true
			case "ck", "check":
				return bcc.bi.Action(domain.BadugiActionCheck, 0, 0), true
			case "a", "allin":
				return bcc.bi.Action(domain.BadugiActionAllIn, 0, 0), true
			case "bl", "bettinglimit":
				return cuiutil.WithParsedIntKeys(args, "bettingLimitTypeRequired0Fixed1Potlimit2Nolimit", "invalidBettingLimit02", 0, 2, func(v int) string {
					cfg := bcc.bi.GetConfig()
					cfg.BettingLimit = domain.BettingLimitType(v)
					return bcc.bi.ResetWithConfig(cfg, nil)
				})
			case "scc", "setcpucount":
				return cuiutil.WithParsedIntKeys(args, "cpuPlayerCountRequired", "invalidCpuPlayerCount13", domain.BadugiCpuCountMin, domain.BadugiCpuCountMax, func(v int) string {
					cfg := bcc.bi.GetConfig()
					cfg.CpuCount = v
					return bcc.bi.ResetWithConfig(cfg, nil)
				})
			case "mai", "metaai":
				if len(args) < 1 {
					return i18n.T("metaAIRequired"), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 1 {
					return invalidArg("invalidMetaAI", "val", args[0]), true
				}
				cfg := bcc.bi.GetConfig()
				cfg.CpuMetaAI = v == 1
				return bcc.bi.ResetWithConfig(cfg, nil), true
			default:
				return handleCuiHintAndLog(cmd, bcc.bi.Hint, bcc.bi.ActionLog)
			}
		},
	)
}

// parseCuiAmount parses the first positive integer argument, returning 0 if
// missing or invalid. Matches PokerCuiController's bet/raise argument handling.
func parseCuiAmount(args []string) int {
	if len(args) == 0 {
		return 0
	}
	a, err := strconv.Atoi(args[0])
	if err != nil || a <= 0 {
		return 0
	}
	return a
}
