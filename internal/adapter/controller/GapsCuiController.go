package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GapsCuiController はGapsゲームのCUIコントローラー。
type GapsCuiController struct {
	gi usecase.GapsInteractorIF
}

// NewGapsCuiController はGapsCuiControllerを生成する。
func NewGapsCuiController(gi usecase.GapsInteractorIF) *GapsCuiController {
	return &GapsCuiController{gi: gi}
}

// Exec はコマンドを実行する。
func (c *GapsCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.gi.Reset()
		},
		[]string{"m", "move", "rd", "redeal", "g", "giveup", "h", "hint", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "m", "move":
				return c.handleMove(args), true
			case "rd", "redeal":
				return c.gi.Redeal(), true
			case "g", "giveup":
				return c.gi.GiveUp(), true
			case "u", "undo":
				return c.gi.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.gi.Hint, c.gi.ActionLog)
			}
		},
	)
}

// handleMove は "m <fromRow> <fromCol> <toRow> <toCol>" を処理する。
func (c *GapsCuiController) handleMove(args []string) string {
	if len(args) != 4 {
		return "Usage: m <fromRow> <fromCol> <toRow> <toCol>"
	}
	fr, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Sprintf("Invalid fromRow: %s.", args[0])
	}
	fc, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Sprintf("Invalid fromCol: %s.", args[1])
	}
	tr, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Sprintf("Invalid toRow: %s.", args[2])
	}
	tc, err := strconv.Atoi(args[3])
	if err != nil {
		return fmt.Sprintf("Invalid toCol: %s.", args[3])
	}
	return c.gi.Move(fr, fc, tr, tc)
}
