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
			case "h", "hint":
				return c.ti.Hint(), true
			case "log", "l":
				return c.ti.ActionLog(), true
			case "u", "undo":
				return c.ti.Undo(), true
			}
			return "", false
		},
	)
}

// handleRemove 除去コマンドを処理
func (c *TriPeaksCuiController) handleRemove(args []string) string {
	if len(args) != 2 {
		return "Usage: rm <row> <col>"
	}
	row, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Sprintf("Invalid row: %s.", args[0])
	}
	col, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Sprintf("Invalid col: %s.", args[1])
	}
	return c.ti.Remove(row, col)
}
