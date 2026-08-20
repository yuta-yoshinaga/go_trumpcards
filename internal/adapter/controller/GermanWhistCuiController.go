//go:build !js || !wasm || classic

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GermanWhistCuiController ジャーマンホイストCUIコントローラークラス
type GermanWhistCuiController struct {
	gi usecase.GermanWhistInteractorIF
}

// NewGermanWhistCuiController コンストラクタ
func NewGermanWhistCuiController(gi usecase.GermanWhistInteractorIF) *GermanWhistCuiController {
	return &GermanWhistCuiController{gi: gi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit      → ゲーム終了 ("bye.")
//	r / reset     → ゲームリセット
//	p / play <i>  → 手札の i 番目を出す
//	g / giveup    → 投了
//	h / hint      → ヒント表示
//	log / l       → 棋譜表示
func (c *GermanWhistCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.gi.Reset() },
		[]string{"p", "play", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.gi.Play)
			case "g", "giveup":
				return c.gi.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.gi.Hint, c.gi.ActionLog)
			}
		},
	)
}
