//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CruelCuiController クルーエルCUIコントローラークラス
type CruelCuiController struct {
	ci usecase.CruelInteractorIF
}

// NewCruelCuiController コンストラクタ
func NewCruelCuiController(ci usecase.CruelInteractorIF) *CruelCuiController {
	return &CruelCuiController{ci: ci}
}

// Exec コマンド実行
func (c *CruelCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.ci.Reset()
		},
		[]string{"m", "move", "s", "shift", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "m", "move":
				return c.handleMove(args), true
			case "s", "shift":
				return c.ci.Shift(), true
			case "g", "giveup":
				return c.ci.GiveUp(), true
			case "ac", "autocomplete":
				return c.ci.AutoComplete(), true
			case "u", "undo":
				return c.ci.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.ci.Hint, c.ci.ActionLog)
			}
		},
	)
}

// handleMove 移動コマンドを処理（Cruel は最上段のみ移動できるので cardIndex なし）
func (c *CruelCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("cruel.promptSourceColumn"), "m {0}")
	}

	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}

	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("cruel.promptToZone"), fmt.Sprintf("m %d {0}", fromCol))
	}

	dest := args[1]
	if dest == "f" {
		return c.ci.MoveTableauToFoundation(fromCol)
	}
	toCol, err := strconv.Atoi(dest)
	if err != nil {
		return i18n.MarkError(i18n.T("cruel.moveUsage"))
	}
	return c.ci.MoveTableauToTableau(fromCol, toCol)
}
