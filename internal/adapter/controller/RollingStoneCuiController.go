//go:build !js || !wasm || extra3

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RollingStoneCuiController ローリングストーンCUIコントローラークラス
type RollingStoneCuiController struct {
	ri usecase.RollingStoneInteractorIF
}

// NewRollingStoneCuiController コンストラクタ
func NewRollingStoneCuiController(ri usecase.RollingStoneInteractorIF) *RollingStoneCuiController {
	return &RollingStoneCuiController{ri: ri}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit      → ゲーム終了 ("bye.")
//	r / reset     → ゲームリセット (設定保持)
//	p / play <i>  → 手札の i 番目を出す
//	u / pickup    → フォローできないので場札を引き取る
//	g / giveup    → 投了
//	h / hint      → ヒント表示
//	log / l       → 棋譜表示
//
// **引き取りは負けに近づく行動です。** 手札が増えるので、勝利条件（先に出し切る）
// から遠ざかります。
func (c *RollingStoneCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ri.ResetWithConfig(c.ri.GetConfig()) },
		[]string{"p", "play", "u", "pickup", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
					cuiutil.NoMin, cuiutil.NoMax, c.ri.Play)
			case "u", "pickup":
				return c.ri.PickUp(), true
			case "g", "giveup":
				return c.ri.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.ri.Hint, c.ri.ActionLog)
			}
		},
	)
}
