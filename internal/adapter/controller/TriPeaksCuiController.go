//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TriPeaksCuiController トリピークスCUIコントローラークラス
type TriPeaksCuiController struct {
	ti usecase.TriPeaksInteractorIF
}

// NewTriPeaksCuiController コンストラクタ
func NewTriPeaksCuiController(ti usecase.TriPeaksInteractorIF) *TriPeaksCuiController {
	return &TriPeaksCuiController{ti: ti}
}

// Exec コマンド実行
func (c *TriPeaksCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.ti.Reset()
		},
		[]string{"d", "draw", "rm", "remove", "g", "giveup", "h", "hint", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				return c.ti.Draw(), true
			case "rm", "remove":
				return c.handleRemove(args), true
			case "g", "giveup":
				return c.ti.GiveUp(), true
			case "u", "undo":
				return c.ti.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.ti.Hint, c.ti.ActionLog)
			}
		},
	)
}

// handleRemove 除去コマンドを処理
func (c *TriPeaksCuiController) handleRemove(args []string) string {
	if len(args) != 2 {
		return invalidArg("usageRmRowCol")
	}
	row, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidRowRaw", "val", fmt.Sprint(args[0]))
	}
	col, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidColRaw", "val", fmt.Sprint(args[1]))
	}
	return c.ti.Remove(row, col)
}
