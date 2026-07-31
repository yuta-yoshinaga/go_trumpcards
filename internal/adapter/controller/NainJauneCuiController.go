//go:build !js || !wasm || extra3

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NainJauneCuiController ル・ナン・ジョーヌCUIコントローラークラス
type NainJauneCuiController struct {
	pi usecase.NainJauneInteractorIF
}

// NewNainJauneCuiController コンストラクタ
func NewNainJauneCuiController(pi usecase.NainJauneInteractorIF) *NainJauneCuiController {
	return &NainJauneCuiController{pi: pi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit          → ゲーム終了 ("bye.")
//	r / reset         → ゲームごとリセット (設定保持)
//	p <i>             → 手札 i を出す
//	n / next          → 次のディールへ進む
//	h / hint          → ヒント表示
//	log / l           → 棋譜表示
func (c *NainJauneCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.pi.GetConfig()
			return c.pi.ResetWithConfig(cfg)
		},
		[]string{"p", "play", "n", "next", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				// 下限を 0 に。NoMin だと `p -1` がそのままドメインまで届く。
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", 0, cuiutil.NoMax, c.pi.Play)
			case "n", "next":
				return c.pi.NextDeal(), true
			default:
				return handleCuiHintAndLog(cmd, c.pi.Hint, c.pi.ActionLog)
			}
		},
	)
}
