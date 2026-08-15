//go:build !js || !wasm || classic

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SlobberhannesCuiController スロバーハンネスCUIコントローラークラス
type SlobberhannesCuiController struct {
	si usecase.SlobberhannesInteractorIF
}

// NewSlobberhannesCuiController コンストラクタ
func NewSlobberhannesCuiController(si usecase.SlobberhannesInteractorIF) *SlobberhannesCuiController {
	return &SlobberhannesCuiController{si: si}
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
func (c *SlobberhannesCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.si.ResetWithConfig(c.si.GetConfig())
		},
		[]string{"p", "play", "n", "next", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.si.Play)
			case "n", "next":
				return c.si.NextRound(), true
			case "g", "giveup":
				return c.si.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.si.Hint, c.si.ActionLog)
			}
		},
	)
}
