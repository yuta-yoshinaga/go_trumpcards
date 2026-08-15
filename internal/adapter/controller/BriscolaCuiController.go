package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BriscolaCuiController ブリスコラCUIコントローラークラス
type BriscolaCuiController struct {
	bi usecase.BriscolaInteractorIF
}

// NewBriscolaCuiController コンストラクタ
func NewBriscolaCuiController(bi usecase.BriscolaInteractorIF) *BriscolaCuiController {
	return &BriscolaCuiController{bi: bi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit         → ゲーム終了 ("bye.")
//	r / reset        → ゲームリセット (設定保持)
//	p / play <i>     → カードをプレイ
//	n / next         → 次のトリックへ
//	h / hint         → ヒント表示
//	log / l          → 棋譜表示
func (c *BriscolaCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.bi.GetConfig()
			return c.bi.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "n", "next", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.bi.Play)
			case "n", "next":
				return c.bi.NextTrick(), true
			default:
				return handleCuiHintAndLog(cmd, c.bi.Hint, c.bi.ActionLog)
			}
		},
	)
}
