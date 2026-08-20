//go:build !js || !wasm || extra2

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ChineseTenCuiController 撿紅點CUIコントローラークラス
type ChineseTenCuiController struct {
	ci usecase.ChineseTenInteractorIF
}

// NewChineseTenCuiController コンストラクタ
func NewChineseTenCuiController(ci usecase.ChineseTenInteractorIF) *ChineseTenCuiController {
	return &ChineseTenCuiController{ci: ci}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit       → ゲーム終了 ("bye.")
//	r / reset      → ゲームリセット (設定保持)
//	p / play <i>   → 手札の札を出す
//	s / select <i> → 取れる場札が複数あるとき、どれを取るか選ぶ
//	h / hint       → ヒント表示
//	log / l        → 棋譜表示
func (c *ChineseTenCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{"p", "play", "s", "select", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Play)
			case "s", "select":
				return cuiutil.WithParsedIntKeys(args, "layoutIndexRequired", "invalidLayoutIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Select)
			default:
				return handleCuiHintAndLog(cmd, c.ci.Hint, c.ci.ActionLog)
			}
		},
	)
}
