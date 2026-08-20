//go:build !js || !wasm || solo

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SchnapsenCuiController シュナプセンCUIコントローラークラス
type SchnapsenCuiController struct {
	si usecase.SchnapsenInteractorIF
}

// NewSchnapsenCuiController コンストラクタ
func NewSchnapsenCuiController(si usecase.SchnapsenInteractorIF) *SchnapsenCuiController {
	return &SchnapsenCuiController{si: si}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit         → ゲーム終了 ("bye.")
//	r / reset        → ゲームリセット (設定保持)
//	p / play <i>     → カードをプレイ
//	m / marriage <i> → マリアージュを宣言してその K/Q をリード
//	n / next         → 次のトリックへ
//	h / hint         → ヒント表示
//	log / l          → 棋譜表示
func (c *SchnapsenCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.si.GetConfig()
			return c.si.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "m", "marriage", "n", "next", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.si.Play)
			case "m", "marriage":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.si.DeclareMarriage)
			case "n", "next":
				return c.si.NextTrick(), true
			default:
				return handleCuiHintAndLog(cmd, c.si.Hint, c.si.ActionLog)
			}
		},
	)
}
