//go:build !js || !wasm || extra2

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PigCuiController ピッグCUIコントローラークラス
type PigCuiController struct {
	pi usecase.PigInteractorIF
}

// NewPigCuiController コンストラクタ
func NewPigCuiController(pi usecase.PigInteractorIF) *PigCuiController {
	return &PigCuiController{pi: pi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit      → ゲーム終了 ("bye.")
//	r / reset     → ゲームリセット (設定保持)
//	p / pass <i>  → 手札の i 番目を左隣へ渡す
//	s / signal    → 合図に気づいたと伝える (**遅れた 1 人が文字をもらう**)
//	n / next      → 次のラウンドを配る
//	g / giveup    → 投了
//	h / hint      → ヒント表示
//	log / l       → 棋譜表示
func (c *PigCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.pi.ResetWithConfig(c.pi.GetConfig()) },
		[]string{"p", "pass", "s", "signal", "n", "next", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "pass":
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.",
					cuiutil.NoMin, cuiutil.NoMax, c.pi.Pass)
			case "s", "signal":
				return c.pi.Signal(), true
			case "n", "next":
				return c.pi.NextRound(), true
			case "g", "giveup":
				return c.pi.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.pi.Hint, c.pi.ActionLog)
			}
		},
	)
}
