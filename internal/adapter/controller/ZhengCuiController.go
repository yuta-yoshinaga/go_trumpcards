//go:build !js || !wasm || solo

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ZhengCuiController 争上游 CUIコントローラークラス
type ZhengCuiController struct {
	zi usecase.ZhengInteractorIF
}

// NewZhengCuiController コンストラクタ
func NewZhengCuiController(zi usecase.ZhengInteractorIF) *ZhengCuiController {
	return &ZhengCuiController{zi: zi}
}

// Exec コマンド実行
func (c *ZhengCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.zi.GetConfig()
			return c.zi.ResetWithConfig(cfg)
		},
		[]string{"p", "play", "sd", "setdifficulty", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				indices, skipped := cuiutil.ParseIntSlice(args)
				return cuiutil.PrependSkippedWarning(c.zi.Play(indices), skipped), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Normal, 1=Easy, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.zi.GetConfig()
					cfg.CpuDifficulty = domain.ZhengCpuDifficulty(v)
					return c.zi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.zi.ActionLog)
			}
		},
	)
}
