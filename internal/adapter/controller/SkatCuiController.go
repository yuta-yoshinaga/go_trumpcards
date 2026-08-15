//go:build !js || !wasm || extra3

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SkatCuiController Skat CUI controller.
type SkatCuiController struct {
	si usecase.SkatInteractorIF
}

// NewSkatCuiController constructs a SkatCuiController.
func NewSkatCuiController(si usecase.SkatInteractorIF) *SkatCuiController {
	return &SkatCuiController{si: si}
}

// Exec dispatches a CUI command.
//
// Commands:
//
//	q / quit                     → quit ("bye.")
//	r / reset                    → reset (config preserved)
//	b / bid <0|1>                → pass (0) or accept (1) the active bid step
//	ps / pickskat <0|1>          → decline (0) or pick up (1) the skat
//	d / discard <i> <j>          → discard the two cards at indices i and j
//	g / game <type> [trumpSuit]  → declare game (1=Suit 2=Grand 3=Null)
//	p / play <i>                 → play card i
//	n / next                     → next trick
//	nr / nextround               → score round and start next
//	sd / setdifficulty <0-2>     → set CPU difficulty
//	sl / settarget <n>           → set game-end target score
//	h / hint                     → show hint
//	log / l                      → show action log
func (c *SkatCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.si.GetConfig()
			return c.si.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bid", "ps", "pickskat", "d", "discard",
			"g", "game", "p", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sl", "settarget",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				return cuiutil.WithParsedIntKeys(args, "bidStepRequired", "invalidBidStep", 0, 1, func(v int) string {
					return c.si.Bid(v == 1)
				})
			case "ps", "pickskat":
				return cuiutil.WithParsedInt(args, "Pickup decision is required (0=decline, 1=pick up).", "Invalid pickup decision: %s.", 0, 1, func(v int) string {
					return c.si.PickSkat(v == 1)
				})
			case "d", "discard":
				if len(args) < 2 {
					return "Usage: discard <i> <j> (two card indices)\n", true
				}
				idxA, errMsg, ok := cuiutil.ParseIntArgKeys(args[:1], "", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax)
				if !ok {
					return errMsg, true
				}
				idxB, errMsg, ok := cuiutil.ParseIntArgKeys(args[1:2], "", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax)
				if !ok {
					return errMsg, true
				}
				return c.si.Discard(idxA, idxB), true
			case "g", "game":
				if len(args) < 1 {
					return "Usage: game <type> [trumpSuit]\n  type: 1=Suit 2=Grand 3=Null\n  trumpSuit (suit only): 1=♠ 2=♣ 3=♥ 4=♦\n", true
				}
				gt, errMsg, ok := cuiutil.ParseIntArg(args[:1], "", "Invalid game type: %s.", 1, 3)
				if !ok {
					return errMsg, true
				}
				trump := 0
				if len(args) >= 2 {
					t, msg, tok := cuiutil.ParseIntArgKeys(args[1:2], "", "invalidTrumpSuit", 1, 4)
					if !tok {
						return msg, true
					}
					trump = t
				}
				return c.si.DeclareGame(domain.SkatGameType(gt), trump), true
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.si.Play)
			case "n", "next":
				return c.si.NextTrick(), true
			case "nr", "nextround":
				return c.si.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.si.GetConfig()
					cfg.CpuDifficulty = domain.SkatCpuDifficulty(v)
					return c.si.ResetWithConfig(cfg)
				})
			case "sl", "settarget":
				return cuiutil.WithParsedIntKeys(args, "targetScoreRequired", "invalidTargetScore", 1, math.MaxInt, func(v int) string {
					cfg := c.si.GetConfig()
					cfg.TargetScore = v
					return c.si.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.si.Hint, c.si.ActionLog)
			}
		},
	)
}
