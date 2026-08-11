//go:build !js || !wasm || extra2

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RamsCuiController ラムスCUIコントローラークラス
type RamsCuiController struct {
	ri usecase.RamsInteractorIF
}

// NewRamsCuiController コンストラクタ
func NewRamsCuiController(ri usecase.RamsInteractorIF) *RamsCuiController {
	return &RamsCuiController{ri: ri}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit      → ゲーム終了 ("bye.")
//	r / reset     → ゲームリセット (設定保持)
//	in / play     → このラウンドに参加する
//	out / pass    → このラウンドは降りる
//	c / card <i>  → 手札の i 番目を出す
//	n / next      → 次のラウンドへ
//	g / giveup    → 投了
//	h / hint      → ヒント表示
//	log / l       → 棋譜表示
func (c *RamsCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ri.ResetWithConfig(c.ri.GetConfig()) },
		[]string{"in", "play", "out", "pass", "c", "card", "n", "next", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "in", "play":
				return c.ri.Play(), true
			case "out", "pass":
				return c.ri.Pass(), true
			case "c", "card":
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax, c.ri.PlayCard)
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
