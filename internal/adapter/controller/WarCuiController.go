package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// WarCuiController 戦争CUIコントローラークラス
type WarCuiController struct {
	wi usecase.WarInteractorIF
}

// NewWarCuiController コンストラクタ
func NewWarCuiController(wi usecase.WarInteractorIF) *WarCuiController {
	return &WarCuiController{wi: wi}
}

// Exec コマンド実行
//
//	q / quit        → ゲーム終了
//	r / reset       → ゲームリセット
//	s / step        → 次の1ステップを進める (めくり/解決/戦争解決)
//	a / autoplay    → 決着まで自動で進める
//	sm / setmax <n> → MaxRounds を設定してリセット
//	log / l         → 棋譜表示
func (c *WarCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.wi.GetConfig()
			return c.wi.ResetWithConfig(cfg)
		},
		[]string{"s", "step", "a", "autoplay", "sm", "setmax", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "s", "step":
				return c.wi.Step(), true
			case "a", "autoplay":
				return c.wi.AutoPlay(), true
			case "sm", "setmax":
				val, errMsg, ok := cuiutil.ParseOptionalIntKeys(args, 0, domain.WarDefaultMaxRounds, "invalidMaxRounds")
				if !ok {
					return errMsg, true
				}
				cfg := c.wi.GetConfig()
				cfg.MaxRounds = val
				return c.wi.ResetWithConfig(cfg), true
			case "log", "l":
				return c.wi.ActionLog(), true
			default:
				return "", false
			}
		},
	)
}
