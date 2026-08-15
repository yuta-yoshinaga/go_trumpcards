//go:build !js || !wasm || extra2

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PolignacCuiController ポリニャックCUIコントローラークラス
type PolignacCuiController struct {
	pi usecase.PolignacInteractorIF
}

// NewPolignacCuiController コンストラクタ
func NewPolignacCuiController(pi usecase.PolignacInteractorIF) *PolignacCuiController {
	return &PolignacCuiController{pi: pi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit      → ゲーム終了 ("bye.")
//	r / reset     → ゲームリセット (設定保持)
//	c / capot     → capot（全トリック獲得）を宣言する
//	pass          → 宣言せずにプレイへ進む
//	p / play <i>  → 手札の i 番目を出す
//	n / next      → 次のラウンドへ
//	g / giveup    → 投了
//	h / hint      → ヒント表示
//	log / l       → 棋譜表示
func (c *PolignacCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.pi.ResetWithConfig(c.pi.GetConfig()) },
		[]string{"c", "capot", "pass", "p", "play", "n", "next", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "c", "capot":
				return c.pi.DeclareCapot(), true
			case "pass":
				return c.pi.Pass(), true
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.pi.Play)
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
