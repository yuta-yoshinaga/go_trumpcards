//go:build !js || !wasm || classic

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ReversisCuiController レヴェルシCUIコントローラークラス
type ReversisCuiController struct {
	ri usecase.ReversisInteractorIF
}

// NewReversisCuiController コンストラクタ
func NewReversisCuiController(ri usecase.ReversisInteractorIF) *ReversisCuiController {
	return &ReversisCuiController{ri: ri}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit      → ゲーム終了 ("bye.")
//	r / reset     → ゲームリセット (設定保持)
//	p / play <i>  → 手札の i 番目を出す
//	n / next      → 次のラウンドへ
//	g / giveup    → 投了
//	h / hint      → ヒント表示
//	log / l       → 棋譜表示
func (c *ReversisCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ri.ResetWithConfig(c.ri.GetConfig()) },
		[]string{"p", "play", "n", "next", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax, c.ri.Play)
			case "n", "next":
				return c.ri.NextRound(), true
			case "g", "giveup":
				return c.ri.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.ri.Hint, c.ri.ActionLog)
			}
		},
	)
}
