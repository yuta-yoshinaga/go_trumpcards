//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GolfCuiController ゴルフソリティアCUIコントローラークラス
type GolfCuiController struct {
	gi usecase.GolfInteractorIF
}

// NewGolfCuiController コンストラクタ
func NewGolfCuiController(gi usecase.GolfInteractorIF) *GolfCuiController {
	return &GolfCuiController{gi: gi}
}

// Exec コマンド実行
func (c *GolfCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.gi.Reset()
		},
		[]string{"d", "draw", "rm", "remove", "g", "giveup", "h", "hint", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				return c.gi.Draw(), true
			case "rm", "remove":
				return c.handleRemove(args), true
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

// handleRemove 除去コマンドを処理
func (c *GolfCuiController) handleRemove(args []string) string {
	if len(args) != 1 {
		return "Usage: rm <col>"
	}
	col, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColRaw", "val", fmt.Sprint(args[0]))
	}
	return c.gi.Remove(col)
}
