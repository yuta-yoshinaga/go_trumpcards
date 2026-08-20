//go:build !js || !wasm || extra3

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SkitgubbeCuiController シートグッベCUIコントローラークラス
type SkitgubbeCuiController struct {
	si usecase.SkitgubbeInteractorIF
}

// NewSkitgubbeCuiController コンストラクタ
func NewSkitgubbeCuiController(si usecase.SkitgubbeInteractorIF) *SkitgubbeCuiController {
	return &SkitgubbeCuiController{si: si}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit     → ゲーム終了 ("bye.")
//	r / reset    → ゲームリセット (設定保持)
//	p / play <i> → 手札の札を出す
//	u / pickup   → 第2フェーズで場の札を引き取る (出せる札があるときは不可)
//	h / hint     → ヒント表示
//	log / l      → 棋譜表示
func (c *SkitgubbeCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.si.GetConfig()
			return c.si.ResetWithConfig(cfg)
		},
		[]string{"p", "play", "u", "pickup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.si.Play)
			case "u", "pickup":
				return c.si.PickUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.si.Hint, c.si.ActionLog)
			}
		},
	)
}
