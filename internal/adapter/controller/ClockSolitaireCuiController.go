//go:build !js || !wasm || solo

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ClockSolitaireCuiController クロックソリティアCUIコントローラークラス
type ClockSolitaireCuiController struct {
	ci usecase.ClockSolitaireInteractorIF
}

// NewClockSolitaireCuiController コンストラクタ
func NewClockSolitaireCuiController(ci usecase.ClockSolitaireInteractorIF) *ClockSolitaireCuiController {
	return &ClockSolitaireCuiController{ci: ci}
}

// Exec コマンド実行
func (c *ClockSolitaireCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.ci.Reset()
		},
		[]string{"s", "step", "a", "autoplay", "u", "undo", "log", "l"},
		func(cmd string, _ []string) (string, bool) {
			switch cmd {
			case "s", "step":
				return c.ci.Step(), true
			case "a", "autoplay":
				return c.ci.AutoPlay(), true
			case "u", "undo":
				return c.ci.Undo(), true
			default:
				return handleCuiLog(cmd, c.ci.ActionLog)
			}
		},
	)
}
