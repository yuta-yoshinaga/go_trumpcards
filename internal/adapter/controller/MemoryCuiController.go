package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MemoryCuiController 神経衰弱CUIコントローラークラス
type MemoryCuiController struct {
	mi usecase.MemoryInteractorIF
}

// NewMemoryCuiController コンストラクタ
func NewMemoryCuiController(mi usecase.MemoryInteractorIF) *MemoryCuiController {
	return &MemoryCuiController{mi: mi}
}

// Exec コマンド実行
func (c *MemoryCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.mi.GetConfig()
			return c.mi.ResetWithConfig(cfg)
		},
		[]string{"f", "flip", "n", "next", "sd", "setdifficulty", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "f", "flip":
				return cuiutil.WithParsedInt(args, "Position is required.", "Invalid position: %s.", cuiutil.NoMin, cuiutil.NoMax, c.mi.Flip)
			case "n", "next":
				return c.mi.Next(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.mi.GetConfig()
					cfg.CpuDifficulty = domain.MemoryCpuDifficulty(v)
					return c.mi.ResetWithConfig(cfg)
				})
			case "log", "l":
				return c.mi.ActionLog(), true
			}
			return "", false
		},
	)
}
