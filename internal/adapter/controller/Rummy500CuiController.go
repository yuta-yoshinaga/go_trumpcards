//go:build !js || !wasm || extra

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// Rummy500CuiController Rummy 500CUIコントローラークラス
type Rummy500CuiController struct {
	ci usecase.Rummy500InteractorIF
}

// NewRummy500CuiController コンストラクタ
func NewRummy500CuiController(ci usecase.Rummy500InteractorIF) *Rummy500CuiController {
	return &Rummy500CuiController{ci: ci}
}

// Exec コマンド実行
func (c *Rummy500CuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"ds", "drawstock", "dd", "drawdiscard",
			"m", "meld", "lo", "layoff", "d", "discard",
			"nr", "nextround",
			"h", "hint",
			"sd", "setdifficulty", "sl", "setlimit", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "ds", "drawstock":
				return c.ci.DrawFromStock(), true
			case "dd", "drawdiscard":
				if len(args) == 0 {
					return c.ci.DrawFromDiscard(-1), true
				}
				return cuiutil.WithParsedIntKeys(args, "discardIndexRequired", "invalidDiscardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.DrawFromDiscard)
			case "m", "meld":
				return c.ci.Meld(parseIntList(args)), true
			case "lo", "layoff":
				indices := parseIntList(args)
				if len(indices) != 3 {
					return "param error: layoff requires 3 ints: <owner> <meldIdx> <cardIdx>.", true
				}
				return c.ci.Layoff(indices[0], indices[1], indices[2]), true
			case "d", "discard":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Discard)
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.Rummy500CpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedInt(args, "Point limit is required.", "Invalid point limit: %s. Please enter 1 or more.", 1, math.MaxInt, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.PointLimit = v
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.ci.Hint, c.ci.ActionLog)
			}
		},
	)
}
