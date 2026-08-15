//go:build !js || !wasm || extra2

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TichuCuiController ティチューCUIコントローラークラス
type TichuCuiController struct {
	tgi usecase.TichuInteractorIF
}

// NewTichuCuiController コンストラクタ
func NewTichuCuiController(tgi usecase.TichuInteractorIF) *TichuCuiController {
	return &TichuCuiController{tgi: tgi}
}

// Exec CUIコマンドを実行する
func (c *TichuCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.tgi.GetConfig()
			return c.tgi.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "d", "declare", "sd", "setdifficulty",
			"log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				indices, skipped := cuiutil.ParseIntSlice(args)
				return cuiutil.PrependSkippedWarning(c.tgi.Play(indices), skipped), true
			case "d", "declare":
				if len(args) == 0 {
					return c.tgi.Declare(0), true
				}
				return cuiutil.WithParsedInt(args, "Declaration is required (0=none, 1=Tichu, 2=Grand Tichu).", "Invalid declaration: %s. Please enter 0-2.", 0, 2, func(v int) string {
					return c.tgi.Declare(v)
				})
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequiredAlt", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.tgi.GetConfig()
					cfg.CpuDifficulty = domain.TichuCpuDifficulty(v)
					return c.tgi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.tgi.ActionLog)
			}
		},
	)
}
