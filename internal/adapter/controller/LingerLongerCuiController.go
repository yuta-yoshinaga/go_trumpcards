//go:build !js || !wasm || extra4

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// LingerLongerCuiController リンガーロンガーCUIコントローラークラス
type LingerLongerCuiController struct {
	li usecase.LingerLongerInteractorIF
}

// NewLingerLongerCuiController コンストラクタ
func NewLingerLongerCuiController(li usecase.LingerLongerInteractorIF) *LingerLongerCuiController {
	return &LingerLongerCuiController{li: li}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit      → ゲーム終了 ("bye.")
//	r / reset     → ゲームリセット (設定保持)
//	p / play <i>  → 手札の i 番目を出す
//	g / giveup    → 投了
//	h / hint      → ヒント表示
//	log / l       → 棋譜表示
//
// **補充するコマンドはありません。** トリックを取れば自動的に 1 枚引きます。
func (c *LingerLongerCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.li.ResetWithConfig(c.li.GetConfig()) },
		[]string{"p", "play", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
					cuiutil.NoMin, cuiutil.NoMax, c.li.Play)
			case "g", "giveup":
				return c.li.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.li.Hint, c.li.ActionLog)
			}
		},
	)
}
