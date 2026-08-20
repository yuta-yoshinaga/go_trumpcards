//go:build !js || !wasm || extra

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BhabhiCuiController バービーCUIコントローラークラス
type BhabhiCuiController struct {
	bi usecase.BhabhiInteractorIF
}

// NewBhabhiCuiController コンストラクタ
func NewBhabhiCuiController(bi usecase.BhabhiInteractorIF) *BhabhiCuiController {
	return &BhabhiCuiController{bi: bi}
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
// **次のハンドへ進むコマンドは無い。** 配り切りの 1 ゲームで最後の 1 人
// (Bhabhi) が決まるまで続くので、ハンドの区切りがありません。
func (c *BhabhiCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.bi.ResetWithConfig(c.bi.GetConfig()) },
		[]string{"p", "play", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.bi.Play)
			case "g", "giveup":
				return c.bi.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.bi.Hint, c.bi.ActionLog)
			}
		},
	)
}
