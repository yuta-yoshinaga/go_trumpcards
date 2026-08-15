//go:build !js || !wasm || extra2

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CuckooCuiController Cuckoo (カッコー) CUIコントローラークラス
type CuckooCuiController struct {
	ci usecase.CuckooInteractorIF
}

// NewCuckooCuiController コンストラクタ
func NewCuckooCuiController(ci usecase.CuckooInteractorIF) *CuckooCuiController {
	return &CuckooCuiController{ci: ci}
}

// Exec コマンド実行
func (c *CuckooCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"k", "keep", "s", "swap",
			"rf", "refuse", "ac", "accept",
			"nr", "nextround",
			"sd", "setdifficulty", "sv", "setlives", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "k", "keep":
				return c.ci.Keep(), true
			case "s", "swap":
				return c.ci.Swap(), true
			case "rf", "refuse":
				return c.ci.Refuse(), true
			case "ac", "accept":
				return c.ci.AcceptSwap(), true
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.CuckooCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sv", "setlives":
				return cuiutil.WithParsedInt(args, "Initial lives is required.", "Invalid lives: %s.", domain.CuckooMinLives, domain.CuckooMaxLives, func(v int) string {
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
