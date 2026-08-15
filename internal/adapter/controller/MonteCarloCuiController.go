//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MonteCarloCuiController はモンテカルロ・ソリティア CUI コントローラー。
type MonteCarloCuiController struct {
	mi usecase.MonteCarloInteractorIF
}

// NewMonteCarloCuiController はコンストラクタ。
func NewMonteCarloCuiController(mi usecase.MonteCarloInteractorIF) *MonteCarloCuiController {
	return &MonteCarloCuiController{mi: mi}
}

// Exec はコマンドを実行する。
func (c *MonteCarloCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.mi.Reset() },
		[]string{"m", "move", "remove", "d", "deal", "u", "undo", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "m", "move", "remove":
				return c.handleRemove(args), true
			case "d", "deal":
				return c.mi.Deal(), true
			case "u", "undo":
				return c.mi.Undo(), true
			case "g", "giveup":
				return c.mi.GiveUp(), true
			case "h", "hint":
				return c.mi.Hint(), true
			default:
				return handleCuiLog(cmd, c.mi.ActionLog)
			}
		},
	)
}

// handleRemove は m コマンドを処理する (引数 4 個必須)。
func (c *MonteCarloCuiController) handleRemove(args []string) string {
	if len(args) != 4 {
		return "Usage: m <r1> <c1> <r2> <c2>"
	}
	parsed := make([]int, 4)
	for i, raw := range args {
		v, err := strconv.Atoi(raw)
		if err != nil {
			return invalidArg("invalidArgumentRaw", "val", fmt.Sprint(raw))
		}
		parsed[i] = v
	}
	return c.mi.Remove(parsed[0], parsed[1], parsed[2], parsed[3])
}
