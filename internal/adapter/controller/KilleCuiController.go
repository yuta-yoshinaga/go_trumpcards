//go:build !js || !wasm || extra3

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KilleCuiController キッレ (Kille) CUIコントローラークラス
type KilleCuiController struct {
	ki usecase.KilleInteractorIF
}

// NewKilleCuiController コンストラクタ
func NewKilleCuiController(ki usecase.KilleInteractorIF) *KilleCuiController {
	return &KilleCuiController{ki: ki}
}

// Exec コマンド実行
func (c *KilleCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ki.GetConfig()
			return c.ki.ResetWithConfig(cfg)
		},
		[]string{
			"e", "exchange", "s", "satisfied",
			"re", "reenter", "nr", "nextround",
			"st", "setstake", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "e", "exchange":
				return c.ki.Exchange(), true
			case "s", "satisfied":
				return c.ki.Satisfied(), true
			case "re", "reenter":
				return c.ki.Reenter(), true
			case "nr", "nextround":
				return c.ki.NextRound(), true
			case "st", "setstake":
				return cuiutil.WithParsedIntKeys(args, "stakeRequired", "invalidStake", 1, 100, func(v int) string {
					cfg := c.ki.GetConfig()
					cfg.Stake = v
					return c.ki.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.ki.ActionLog)
			}
		},
	)
}
