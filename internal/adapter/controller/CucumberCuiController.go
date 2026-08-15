//go:build !js || !wasm || classic

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CucumberCuiController キューカンバーCUIコントローラークラス
type CucumberCuiController struct {
	ci usecase.CucumberInteractorIF
}

// NewCucumberCuiController コンストラクタ
func NewCucumberCuiController(ci usecase.CucumberInteractorIF) *CucumberCuiController {
	return &CucumberCuiController{ci: ci}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit      → ゲーム終了 ("bye.")
//	r / reset     → ゲームリセット (設定保持)
//	p / play <i>  → 手札の i 番目を出す
//	n / next      → 次のラウンドを配る
//	g / giveup    → 投了
//	h / hint      → ヒント表示
//	log / l       → 棋譜表示
func (c *CucumberCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ci.ResetWithConfig(c.ci.GetConfig()) },
		[]string{"p", "play", "n", "next", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
					cuiutil.NoMin, cuiutil.NoMax, c.ci.Play)
			case "n", "next":
				return c.ci.NextRound(), true
			case "g", "giveup":
				return c.ci.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.ci.Hint, c.ci.ActionLog)
			}
		},
	)
}
