//go:build !js || !wasm || extra2

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SpoonsCuiController はスプーンの CUI コントローラー。
type SpoonsCuiController struct {
	si usecase.SpoonsInteractorIF
}

// NewSpoonsCuiController はコンストラクタ。
func NewSpoonsCuiController(si usecase.SpoonsInteractorIF) *SpoonsCuiController {
	return &SpoonsCuiController{si: si}
}

// Exec はコマンドを実行する。
//
//	q / quit        → ゲーム終了
//	r / reset       → ゲームリセット
//	p <n> / pass <n>→ 手札 n 番を次のプレイヤーへ渡す
//	g / grab        → スプーンを掴む
//	n / next        → 次のラウンドへ進む
//	log / l         → 棋譜表示
func (c *SpoonsCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.si.Reset() },
		[]string{"p", "pass", "g", "grab", "n", "next", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "pass":
				// No argument keeps the index-0 shorthand; a typed argument that does
				// not parse is refused rather than silently read as 0, which played a
				// card the player never chose (issue #5390).
				idx := 0
				if len(args) > 0 {
					v, err := strconv.Atoi(args[0])
					if err != nil {
						return invalidArg("invalidCardIndex", "val", args[0]), true
					}
					idx = v
				}
				return c.si.Pass(idx), true
			case "g", "grab":
				return c.si.Grab(), true
			case "n", "next":
				return c.si.NextRound(), true
			case "log", "l":
				return c.si.ActionLog(), true
			default:
				return "", false
			}
		},
	)
}
