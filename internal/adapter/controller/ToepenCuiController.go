//go:build !js || !wasm || extra3

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ToepenCuiController トゥーペンCUIコントローラークラス
type ToepenCuiController struct {
	ti usecase.ToepenInteractorIF
}

// NewToepenCuiController コンストラクタ
func NewToepenCuiController(ti usecase.ToepenInteractorIF) *ToepenCuiController {
	return &ToepenCuiController{ti: ti}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit     → ゲーム終了 ("bye.")
//	r / reset    → ゲームリセット (設定保持)
//	p / play <i> → 手札の札を出す
//	t / toep     → 賭け点を吊り上げる
//	s / stay     → toep に追随する
//	f / fold     → toep に降りる
//	d / redeal   → 貧民 (A/K/Q/J のみ) の手札を捨てて配り直す
//	n / next     → 次のハンドへ
//	h / hint     → ヒント表示
//	log / l      → 棋譜表示
func (c *ToepenCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ti.GetConfig()
			return c.ti.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "t", "toep", "s", "stay", "f", "fold", "d", "redeal", "n", "next", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax, c.ti.Play)
			case "t", "toep":
				return c.ti.Toep(), true
			case "s", "stay":
				return c.ti.Respond(true), true
			case "f", "fold":
				return c.ti.Respond(false), true
			case "d", "redeal":
				return c.ti.Redeal(), true
			case "n", "next":
				return c.ti.NextHand(), true
			default:
				return handleCuiHintAndLog(cmd, c.ti.Hint, c.ti.ActionLog)
			}
		},
	)
}
