//go:build !js || !wasm || extra4

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NarcoticCuiController ナルコティックCUIコントローラークラス
type NarcoticCuiController struct {
	ai usecase.NarcoticInteractorIF
}

// NewNarcoticCuiController コンストラクタ
func NewNarcoticCuiController(ai usecase.NarcoticInteractorIF) *NarcoticCuiController {
	return &NarcoticCuiController{ai: ai}
}

// Exec コマンド実行
func (c *NarcoticCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.ai.Reset()
		},
		[]string{"d", "draw", "rm", "remove", "mv", "move", "rd", "redeal", "g", "giveup", "h", "hint", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				return c.ai.Draw(), true
			case "rm", "remove":
				// **列を取らない。**揃った4枚をまとめて捨てるので、選ぶ余地が無い。
				return c.ai.Remove(), true
			case "rd", "redeal":
				return c.ai.Redeal(), true
			case "mv", "move":
				return c.handleColCommand(args, "mv", c.ai.Move), true
			case "g", "giveup":
				return c.ai.GiveUp(), true
			case "u", "undo":
				return c.ai.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.ai.Hint, c.ai.ActionLog)
			}
		},
	)
}

// handleColCommand 列番号を1つ取るコマンドを処理する。
func (c *NarcoticCuiController) handleColCommand(args []string, name string, fn func(int) string) string {
	// **エラーだけ英語で混ざらないようにする。**presenter 側は丁寧に i18n を
	// 通しているのに、ここだけ英語リテラルを返していた (#4803)。
	if len(args) != 1 {
		return i18n.Tf("narcotic.usageCol", "cmd", name)
	}
	col, err := strconv.Atoi(args[0])
	if err != nil {
		// 列番号のエラーは共通キーがある。ゲームごとに文言を割らない。
		return invalidArg("invalidColumn", "val", args[0])
	}
	return fn(col)
}
