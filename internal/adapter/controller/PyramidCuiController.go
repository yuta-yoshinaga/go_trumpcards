//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PyramidCuiController ピラミッドCUIコントローラークラス
type PyramidCuiController struct {
	pi usecase.PyramidInteractorIF
}

// NewPyramidCuiController コンストラクタ
func NewPyramidCuiController(pi usecase.PyramidInteractorIF) *PyramidCuiController {
	return &PyramidCuiController{pi: pi}
}

// Exec コマンド実行
func (c *PyramidCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.pi.Reset()
		},
		[]string{"d", "draw", "rm", "remove", "g", "giveup", "h", "hint", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				return c.pi.Draw(), true
			case "rm", "remove":
				return c.handleRemove(args), true
			case "g", "giveup":
				return c.pi.GiveUp(), true
			case "u", "undo":
				return c.pi.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.pi.Hint, c.pi.ActionLog)
			}
		},
	)
}

// handleRemove 除去コマンドを処理
func (c *PyramidCuiController) handleRemove(args []string) string {
	if len(args) == 0 {
		return "Usage: rm <row> <col> | rm <r1> <c1> <r2> <c2> | rm w <row> <col> | rm w"
	}

	// rm w ... → ウェイスト関連
	if args[0] == "w" {
		if len(args) == 1 {
			// rm w → ウェイストのキング除去
			return c.pi.RemoveWasteKing()
		}
		if len(args) == 3 {
			// rm w <row> <col> → ウェイスト+ピラミッド
			row, err := strconv.Atoi(args[1])
			if err != nil {
				return invalidArg("invalidRowRaw", "val", fmt.Sprint(args[1]))
			}
			col, err := strconv.Atoi(args[2])
			if err != nil {
				return invalidArg("invalidColRaw", "val", fmt.Sprint(args[2]))
			}
			return c.pi.RemoveWithWaste(row, col)
		}
		return "Usage: rm w | rm w <row> <col>"
	}

	// rm <row> <col> → キング除去
	if len(args) == 2 {
		row, err := strconv.Atoi(args[0])
		if err != nil {
			return invalidArg("invalidRowRaw", "val", fmt.Sprint(args[0]))
		}
		col, err := strconv.Atoi(args[1])
		if err != nil {
			return invalidArg("invalidColRaw", "val", fmt.Sprint(args[1]))
		}
		return c.pi.RemoveKing(row, col)
	}

	// rm <r1> <c1> <r2> <c2> → ペア除去
	if len(args) == 4 {
		r1, err := strconv.Atoi(args[0])
		if err != nil {
			return invalidArg("invalidRow1", "val", fmt.Sprint(args[0]))
		}
		c1, err := strconv.Atoi(args[1])
		if err != nil {
			return invalidArg("invalidCol1", "val", fmt.Sprint(args[1]))
		}
		r2, err := strconv.Atoi(args[2])
		if err != nil {
			return invalidArg("invalidRow2", "val", fmt.Sprint(args[2]))
		}
		c2, err := strconv.Atoi(args[3])
		if err != nil {
			return invalidArg("invalidCol2", "val", fmt.Sprint(args[3]))
		}
		return c.pi.RemovePair(r1, c1, r2, c2)
	}

	return "Usage: rm <row> <col> | rm <r1> <c1> <r2> <c2> | rm w <row> <col> | rm w"
}
