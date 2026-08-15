//go:build !js || !wasm || solo

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ThirtyOneCuiController ThirtyOne CUIコントローラークラス
type ThirtyOneCuiController struct {
	ci usecase.ThirtyOneInteractorIF
}

// NewThirtyOneCuiController コンストラクタ
func NewThirtyOneCuiController(ci usecase.ThirtyOneInteractorIF) *ThirtyOneCuiController {
	return &ThirtyOneCuiController{ci: ci}
}

// Exec コマンド実行
func (c *ThirtyOneCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"ds", "drawstock", "dd", "drawdiscard", "d", "discard",
			"k", "knock",
			"nr", "nextround",
			"sd", "setdifficulty", "sv", "setlives", "log", "l",
			"h", "hint",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "ds", "drawstock":
				return c.ci.DrawFromStock(), true
			case "dd", "drawdiscard":
				return c.ci.DrawFromDiscard(), true
			case "d", "discard":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Discard)
			case "k", "knock":
				return c.ci.Knock(), true
			case "h", "hint":
				return c.ci.Hint(), true
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.ThirtyOneCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sv", "setlives":
				return cuiutil.WithParsedInt(args, "Initial lives is required.", "Invalid lives: %s.", domain.ThirtyOneMinLives, domain.ThirtyOneMaxLives, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.InitialLives = v
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.ci.ActionLog)
			}
		},
	)
}
