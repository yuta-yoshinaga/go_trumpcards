//go:build !js || !wasm || solo

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// YanivCuiController Yaniv CUIコントローラークラス
type YanivCuiController struct {
	ci usecase.YanivInteractorIF
}

// NewYanivCuiController コンストラクタ
func NewYanivCuiController(ci usecase.YanivInteractorIF) *YanivCuiController {
	return &YanivCuiController{ci: ci}
}

// Exec コマンド実行
func (c *YanivCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"d", "discard", "y", "yaniv",
			"ds", "drawstock", "dp", "drawpickup",
			"nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "discard":
				indices, skipped := cuiutil.ParseIntSlice(args)
				return cuiutil.PrependSkippedWarning(c.ci.Discard(indices), skipped), true
			case "y", "yaniv":
				return c.ci.DeclareYaniv(), true
			case "ds", "drawstock":
				return c.ci.DrawFromStock(), true
			case "dp", "drawpickup":
				return cuiutil.WithParsedInt(args, "Pickup end is required (0=first, 1=last).", "Invalid pickup end: %s. Please enter 0 or 1.", 0, 1, c.ci.DrawFromPickup)
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.YanivCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedInt(args, "Score limit is required.", "Invalid score limit: %s.", domain.YanivMinScoreLimit, domain.YanivMaxScoreLimit, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.ScoreLimit = v
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.ci.ActionLog)
			}
		},
	)
}
