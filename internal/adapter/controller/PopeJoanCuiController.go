//go:build !js || !wasm || extra3

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PopeJoanCuiController ポープ・ジョーンCUIコントローラークラス
type PopeJoanCuiController struct {
	pi usecase.PopeJoanInteractorIF
}

// NewPopeJoanCuiController コンストラクタ
func NewPopeJoanCuiController(pi usecase.PopeJoanInteractorIF) *PopeJoanCuiController {
	return &PopeJoanCuiController{pi: pi}
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
func (c *PopeJoanCuiController) Exec(command string) string {
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
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", 0, cuiutil.NoMax, c.pi.Play)
			case "n", "next":
				return c.pi.NextDeal(), true
			default:
				return handleCuiHintAndLog(cmd, c.pi.Hint, c.pi.ActionLog)
			}
		},
	)
}
