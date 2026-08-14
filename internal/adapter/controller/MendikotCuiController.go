//go:build !js || !wasm || extra

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MendikotCuiController メンディコットCUIコントローラークラス
type MendikotCuiController struct {
	mi usecase.MendikotInteractorIF
}

// NewMendikotCuiController コンストラクタ
func NewMendikotCuiController(mi usecase.MendikotInteractorIF) *MendikotCuiController {
	return &MendikotCuiController{mi: mi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit      → ゲーム終了 ("bye.")
//	r / reset     → ゲームリセット (設定保持)
//	p / play <i>  → 手札の i 番目を出す
//	n / next      → 次のハンドへ
//	g / giveup    → 投了
//	h / hint      → ヒント表示
//	log / l       → 棋譜表示
//
// **切り札を選ぶコマンドは無い。** フォローできなかったときに出した札の
// スートが、そのまま切り札になる。
func (c *MendikotCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.mi.ResetWithConfig(c.mi.GetConfig()) },
		[]string{"p", "play", "n", "next", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax, c.mi.Play)
			case "n", "next":
				return c.mi.NextHand(), true
			case "g", "giveup":
				return c.mi.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.mi.Hint, c.mi.ActionLog)
			}
		},
	)
}
