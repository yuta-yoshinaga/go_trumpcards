//go:build !js || !wasm || extra3

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PochCuiController ポッホCUIコントローラークラス
type PochCuiController struct {
	pi usecase.PochInteractorIF
}

// NewPochCuiController コンストラクタ
func NewPochCuiController(pi usecase.PochInteractorIF) *PochCuiController {
	return &PochCuiController{pi: pi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit          → ゲーム終了 ("bye.")
//	r / reset         → ゲームごとリセット (設定保持)
//	b / bet           → pochen で1単位賭ける
//	f / fold          → pochen で降りる
//	p <i>             → 手札 i を出す (ストップス)
//	n / next          → 次のディールへ進む
//	h / hint          → ヒント表示
//	log / l           → 棋譜表示
func (c *PochCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.pi.GetConfig()
			return c.pi.ResetWithConfig(cfg)
		},
		[]string{"b", "bet", "f", "fold", "p", "play", "n", "next", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				return c.pi.Bet(), true
			case "f", "fold":
				return c.pi.Fold(), true
			case "p", "play":
				// 下限を 0 にしておく。NoMin だと `p -1` がそのまま
				// ドメインまで届き、同じ拒否をもう一段深いところでやることになる。
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", 0, cuiutil.NoMax, c.pi.Play)
			case "n", "next":
				return c.pi.NextDeal(), true
			default:
				return handleCuiHintAndLog(cmd, c.pi.Hint, c.pi.ActionLog)
			}
		},
	)
}
