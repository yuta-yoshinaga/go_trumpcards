//go:build !js || !wasm || extra2

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CribbageSquaresCuiController はクリベッジ・スクエアズ CUI コントローラー。
type CribbageSquaresCuiController struct {
	pi usecase.CribbageSquaresInteractorIF
}

// NewCribbageSquaresCuiController はコンストラクタ。
func NewCribbageSquaresCuiController(pi usecase.CribbageSquaresInteractorIF) *CribbageSquaresCuiController {
	return &CribbageSquaresCuiController{pi: pi}
}

// Exec はコマンドを実行する。
func (c *CribbageSquaresCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.pi.Reset() },
		[]string{"p", "place", "u", "undo", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "place":
				return c.handlePlace(args), true
			case "u", "undo":
				return c.pi.Undo(), true
			case "g", "giveup":
				return c.pi.GiveUp(), true
			case "h", "hint":
				return c.pi.Hint(), true
			default:
				return handleCuiLog(cmd, c.pi.ActionLog)
			}
		},
	)
}

// handlePlace は place コマンドを処理する。
func (c *CribbageSquaresCuiController) handlePlace(args []string) string {
	if len(args) != 2 {
		return "Usage: p <row> <col>"
	}
	row, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Sprintf("Invalid row: %s.", args[0])
	}
	col, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Sprintf("Invalid col: %s.", args[1])
	}
	return c.pi.Place(row, col)
}
