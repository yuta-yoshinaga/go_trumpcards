//go:build !js || !wasm || extra3

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TrexCuiController トリックスCUIコントローラークラス
type TrexCuiController struct {
	ti usecase.TrexInteractorIF
}

// NewTrexCuiController コンストラクタ
func NewTrexCuiController(ti usecase.TrexInteractorIF) *TrexCuiController {
	return &TrexCuiController{ti: ti}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit     → ゲーム終了 ("bye.")
//	r / reset    → ゲームごとリセット (設定保持)
//	c / choose <n> → 王が契約 n を選ぶ (0=♥K, 1=♦, 2=Q, 3=トリック, 4=ドミノ)
//	p / play <i> → 手札の札を出す
//	s / pass     → ドミノでパスする
//	n / next     → 次のディールへ進む
//	h / hint     → ヒント表示
//	log / l      → 棋譜表示
func (c *TrexCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ti.GetConfig()
			return c.ti.ResetWithConfig(cfg)
		},
		[]string{"c", "choose", "p", "play", "s", "pass", "n", "next", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "c", "choose":
				return cuiutil.WithParsedInt(args, "Contract number is required.", "Invalid contract: %s.", cuiutil.NoMin, cuiutil.NoMax, c.ti.Choose)
			case "p", "play":
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax, c.ti.Play)
			case "s", "pass":
				return c.ti.Pass(), true
			case "n", "next":
				return c.ti.NextDeal(), true
			default:
				return handleCuiHintAndLog(cmd, c.ti.Hint, c.ti.ActionLog)
			}
		},
	)
}
