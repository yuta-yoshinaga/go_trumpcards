//go:build !js || !wasm || solo

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// AcesUpCuiController エースアップCUIコントローラークラス
type AcesUpCuiController struct {
	ai usecase.AcesUpInteractorIF
}

// NewAcesUpCuiController コンストラクタ
func NewAcesUpCuiController(ai usecase.AcesUpInteractorIF) *AcesUpCuiController {
	return &AcesUpCuiController{ai: ai}
}

// Exec コマンド実行
func (c *AcesUpCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.ai.Reset()
		},
		[]string{"d", "draw", "rm", "remove", "mv", "move", "g", "giveup", "h", "hint", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				return c.ai.Draw(), true
			case "rm", "remove":
				return c.handleColCommand(args, "rm", c.ai.Remove), true
			case "mv", "move":
				return c.handleColCommand(args, "mv", c.ai.Move), true
			case "g", "giveup":
				return c.ai.GiveUp(), true
			case "u", "undo":
				return c.ai.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.ai.Hint, c.ai.ActionLog)
			}
		},
	)
}

// handleColCommand 列番号を1つ取るコマンドを処理する。
func (c *AcesUpCuiController) handleColCommand(args []string, name string, fn func(int) string) string {
	// **エラーだけ英語で混ざらないようにする。**presenter 側は丁寧に i18n を
	// 通しているのに、ここだけ英語リテラルを返していた (#4803)。
	if len(args) != 1 {
		return i18n.Tf("acesup.usageCol", "cmd", name)
	}
	col, msg, ok := cuiutil.ParseIntArgKeys(args, "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
	if !ok {
		return msg
	}
	return fn(col)
}
