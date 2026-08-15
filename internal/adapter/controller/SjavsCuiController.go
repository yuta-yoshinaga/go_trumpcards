//go:build !js || !wasm || extra2

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SjavsCuiController シャウスCUIコントローラークラス
type SjavsCuiController struct {
	si usecase.SjavsInteractorIF
}

// NewSjavsCuiController コンストラクタ
func NewSjavsCuiController(si usecase.SjavsInteractorIF) *SjavsCuiController {
	return &SjavsCuiController{si: si}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit     → ゲーム終了 ("bye.")
//	r / reset    → ラバーごとリセット (設定保持)
//	b / bid <n>  → 切札スート長 n を申告する (0 でパス)
//	p / play <i> → 手札の札を出す
//	n / next     → 次のハンドへ進む
//	h / hint     → ヒント表示
//	log / l      → 棋譜表示
func (c *SjavsCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.si.GetConfig()
			return c.si.ResetWithConfig(cfg)
		},
		[]string{"b", "bid", "p", "play", "n", "next", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				return cuiutil.WithParsedIntKeys(args, "bidLengthRequired", "invalidBidLength", cuiutil.NoMin, cuiutil.NoMax, c.si.Bid)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.si.Play)
			case "n", "next":
				return c.si.NextHand(), true
			default:
				return handleCuiHintAndLog(cmd, c.si.Hint, c.si.ActionLog)
			}
		},
	)
}
