//go:build !js || !wasm || extra2

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MushiCuiController 虫CUIコントローラークラス
type MushiCuiController struct {
	mi usecase.MushiInteractorIF
}

// NewMushiCuiController コンストラクタ
func NewMushiCuiController(mi usecase.MushiInteractorIF) *MushiCuiController {
	return &MushiCuiController{mi: mi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit       → ゲーム終了 ("bye.")
//	r / reset      → ゲームリセット (設定保持)
//	p / play <i>   → 手札の札を出す
//	s / select <i> → 同月が 2 枚あるとき、取る場札を選ぶ
//	n / next       → 次のラウンドへ
//	h / hint       → ヒント表示
//	log / l        → 棋譜表示
func (c *MushiCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.mi.GetConfig()
			return c.mi.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "s", "select", "n", "next", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.mi.Play)
			case "s", "select":
				return cuiutil.WithParsedIntKeys(args, "fieldIndexRequired", "invalidFieldIndex", cuiutil.NoMin, cuiutil.NoMax, c.mi.Select)
			case "n", "next":
				return c.mi.NextRound(), true
			default:
				return handleCuiHintAndLog(cmd, c.mi.Hint, c.mi.ActionLog)
			}
		},
	)
}
