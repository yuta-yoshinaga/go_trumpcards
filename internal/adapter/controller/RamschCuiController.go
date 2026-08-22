//go:build !js || !wasm || extra3

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RamschCuiController Ramsch CUI controller.
type RamschCuiController struct {
	si usecase.RamschInteractorIF
}

// NewRamschCuiController constructs a RamschCuiController.
func NewRamschCuiController(si usecase.RamschInteractorIF) *RamschCuiController {
	return &RamschCuiController{si: si}
}

// Exec dispatches a CUI command.
//
// Commands:
//
//	q / quit                     → quit ("bye.")
//	r / reset                    → reset (config preserved)
//	p / play <i>                 → play card i
//	n / next                     → next trick
//	nr / nextround               → score round and start next
//	sd / setdifficulty <0-2>     → set CPU difficulty
//	sl / settarget <n>           → set game-end target score
//	h / hint                     → show hint
//	log / l                      → show action log
func (c *RamschCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.si.GetConfig()
			return c.si.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sl", "settarget",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.si.Play)
			case "n", "next":
				return c.si.NextTrick(), true
			case "nr", "nextround":
				return c.si.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.si.GetConfig()
					cfg.CpuDifficulty = domain.RamschCpuDifficulty(v)
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
