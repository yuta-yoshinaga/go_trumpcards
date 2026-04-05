package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PigsTailCuiController ぶたのしっぽCUIコントローラークラス
type PigsTailCuiController struct {
	pti usecase.PigsTailInteractorIF
}

// NewPigsTailCuiController コンストラクタ
func NewPigsTailCuiController(pti usecase.PigsTailInteractorIF) *PigsTailCuiController {
	return &PigsTailCuiController{pti: pti}
}

// Exec コマンド実行
func (c *PigsTailCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.pti.Reset(c.pti.GetConfig()) },
		[]string{"a", "action", "log", "l"},
		func(cmd string, _ []string) (string, bool) {
			switch cmd {
			case "a", "action":
				return c.pti.Action(0), true
			default:
				return handleCuiLog(cmd, c.pti.ActionLog)
			}
		},
	)
}
