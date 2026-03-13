package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuimsg"
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
		unknownCommandMessage,
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "f", "flip":
				if len(args) < 1 {
					return cuimsg.Required("Position"), true
				}
				pos, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Sprintf("Invalid position: %s.", args[0]), true
				}
				return c.mi.Flip(pos), true
			case "n", "next":
				return c.mi.Next(), true
			case "sd", "setdifficulty":
				if len(args) < 1 {
					return "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 2 {
					return cuimsg.InvalidOutOfRange("CPU difficulty", args[0], "Please enter 0-2."), true
				}
				cfg := c.mi.GetConfig()
				cfg.CpuDifficulty = domain.MemoryCpuDifficulty(v)
				return c.mi.ResetWithConfig(cfg), true
			case "log", "l":
				return c.mi.ActionLog(), true
			}
			return "", false
		},
	)
}
