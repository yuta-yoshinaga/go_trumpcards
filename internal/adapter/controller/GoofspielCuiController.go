//go:build !js || !wasm || extra

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GoofspielCuiController ゴフスピールCUIコントローラークラス
type GoofspielCuiController struct {
	gi usecase.GoofspielInteractorIF
}

// NewGoofspielCuiController コンストラクタ
func NewGoofspielCuiController(gi usecase.GoofspielInteractorIF) *GoofspielCuiController {
	return &GoofspielCuiController{gi: gi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit      → ゲーム終了 ("bye.")
//	r / reset     → ゲームリセット (設定保持)
//	b / bid <i>   → 手札の i 番目を伏せて入札する
//	n / next      → 次の賞札をめくる
//	g / giveup    → 投了
//	h / hint      → ヒント表示
//	log / l       → 棋譜表示
func (c *GoofspielCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.gi.ResetWithConfig(c.gi.GetConfig()) },
		[]string{"b", "bid", "n", "next", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
					cuiutil.NoMin, cuiutil.NoMax, c.gi.Bid)
			case "n", "next":
				return c.gi.NextRound(), true
			case "g", "giveup":
				return c.gi.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.gi.Hint, c.gi.ActionLog)
			}
		},
	)
}
