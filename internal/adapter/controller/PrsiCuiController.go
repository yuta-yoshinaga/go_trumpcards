package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PrsiCuiController プルシーCUIコントローラークラス
type PrsiCuiController struct {
	ci usecase.PrsiInteractorIF
}

// NewPrsiCuiController コンストラクタ
func NewPrsiCuiController(ci usecase.PrsiInteractorIF) *PrsiCuiController {
	return &PrsiCuiController{ci: ci}
}

// Exec コマンド実行
func (c *PrsiCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "d", "draw",
			"sd", "setdifficulty", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Play)
			case "d", "draw":
				return c.ci.Draw(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.PrsiCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.ci.ActionLog)
			}
		},
	)
}
