//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PokerSquaresCuiController はポーカー・スクエアズ CUI コントローラー。
type PokerSquaresCuiController struct {
	pi usecase.PokerSquaresInteractorIF
}

// NewPokerSquaresCuiController はコンストラクタ。
func NewPokerSquaresCuiController(pi usecase.PokerSquaresInteractorIF) *PokerSquaresCuiController {
	return &PokerSquaresCuiController{pi: pi}
}

// Exec はコマンドを実行する。
func (c *PokerSquaresCuiController) Exec(command string) string {
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
func (c *PokerSquaresCuiController) handlePlace(args []string) string {
	if len(args) != 2 {
		return invalidArg("usagePRowCol")
	}
	row, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidRowRaw", "val", fmt.Sprint(args[0]))
	}
	col, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidColRaw", "val", fmt.Sprint(args[1]))
	}
	return c.pi.Place(row, col)
}
