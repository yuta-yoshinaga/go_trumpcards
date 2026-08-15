//go:build !js || !wasm || casino

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BourreCuiController ブーレCUIコントローラークラス
type BourreCuiController struct {
	bgi usecase.BourreInteractorIF
}

// NewBourreCuiController コンストラクタ
func NewBourreCuiController(bgi usecase.BourreInteractorIF) *BourreCuiController {
	return &BourreCuiController{bgi: bgi}
}

// Exec CUIコマンドを実行する
func (c *BourreCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.bgi.GetConfig()
			return c.bgi.ResetWithConfig(cfg)
		},
		[]string{
			"d", "decide", "dr", "draw", "p", "play", "n", "next",
			"sd", "setdifficulty", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "decide":
				return cuiutil.WithParsedInt(args, "Decision is required (1=play, 0=fold).", "Invalid decision: %s. Please enter 0 or 1.", 0, 1, func(v int) string {
					return c.bgi.Decide(v == 1)
				})
			case "dr", "draw":
				indices, skipped := cuiutil.ParseIntSlice(args)
				return cuiutil.PrependSkippedWarning(c.bgi.Draw(indices), skipped), true
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", 0, 4, func(v int) string {
					return c.bgi.Play(v)
				})
			case "n", "next":
				return c.bgi.NextHand(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Normal, 1=Easy, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.bgi.GetConfig()
					cfg.CpuDifficulty = domain.BourreCpuDifficulty(v)
					return c.bgi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.bgi.ActionLog)
			}
		},
	)
}
