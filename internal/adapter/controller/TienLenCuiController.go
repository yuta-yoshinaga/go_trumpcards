//go:build !js || !wasm || solo

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TienLenCuiController Tien Len CUIコントローラークラス
type TienLenCuiController struct {
	tli usecase.TienLenInteractorIF
}

// NewTienLenCuiController コンストラクタ
func NewTienLenCuiController(tli usecase.TienLenInteractorIF) *TienLenCuiController {
	return &TienLenCuiController{tli: tli}
}

// Exec コマンド実行
func (c *TienLenCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.tli.GetConfig()
			return c.tli.ResetWithConfig(cfg)
		},
		[]string{"p", "play", "sd", "setdifficulty", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				indices, skipped := cuiutil.ParseIntSlice(args)
				return cuiutil.PrependSkippedWarning(c.tli.Play(indices), skipped), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequiredAlt", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.tli.GetConfig()
					cfg.CpuDifficulty = domain.TienLenCpuDifficulty(v)
					return c.tli.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.tli.ActionLog)
			}
		},
	)
}
