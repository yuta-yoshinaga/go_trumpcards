//go:build !js || !wasm || extra3

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TarabishCuiController タラビッシュCUIコントローラークラス
type TarabishCuiController struct {
	ti usecase.TarabishInteractorIF
}

// NewTarabishCuiController コンストラクタ
func NewTarabishCuiController(ti usecase.TarabishInteractorIF) *TarabishCuiController {
	return &TarabishCuiController{ti: ti}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit      → ゲーム終了 ("bye.")
//	r / reset     → ゲームリセット (設定保持)
//	t / take      → 表向きの札のスートを切り札として引き受ける
//	pass          → 切り札を見送る（親は見送れない）
//	p / play <i>  → 手札の i 番目を出す
//	n / next      → 次のラウンドへ
//	g / giveup    → 投了
//	h / hint      → ヒント表示
//	log / l       → 棋譜表示
func (c *TarabishCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ti.ResetWithConfig(c.ti.GetConfig()) },
		[]string{"t", "take", "pass", "p", "play", "n", "next", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "t", "take":
				return c.ti.TakeTrump(), true
			case "pass":
				return c.ti.PassTrump(), true
			case "p", "play":
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax, c.ti.Play)
			case "n", "next":
				return c.ti.NextRound(), true
			case "g", "giveup":
				return c.ti.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.ti.Hint, c.ti.ActionLog)
			}
		},
	)
}
