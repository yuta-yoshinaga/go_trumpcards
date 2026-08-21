//go:build !js || !wasm || extra4

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FourteenOutCuiController はフォーティーンアウト・ソリティア CUI コントローラー。
type FourteenOutCuiController struct {
	mi usecase.FourteenOutInteractorIF
}

// NewFourteenOutCuiController はコンストラクタ。
func NewFourteenOutCuiController(mi usecase.FourteenOutInteractorIF) *FourteenOutCuiController {
	return &FourteenOutCuiController{mi: mi}
}

// Exec はコマンドを実行する。
func (c *FourteenOutCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.mi.Reset() },
		[]string{"m", "move", "remove", "u", "undo", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "m", "move", "remove":
				return c.handleRemove(args), true
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

// handleRemove は m コマンドを処理する (列番号 2 個必須)。
//
// **クローン元の Monte Carlo は (行,列) を 4 つ取る。**Fourteen Out で動かせるのは
// 各列の末尾だけなので、列番号 2 つで一意に決まる。
func (c *FourteenOutCuiController) handleRemove(args []string) string {
	if len(args) != 2 {
		return invalidArg("fourteenout.usageMC1C2")
	}
	parsed := make([]int, 2)
	for i, raw := range args {
		v, err := strconv.Atoi(raw)
		if err != nil {
			return invalidArg("invalidArgumentRaw", "val", fmt.Sprint(raw))
		}
		parsed[i] = v
	}
	return c.mi.Remove(parsed[0], parsed[1])
}
