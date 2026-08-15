//go:build !js || !wasm || extra

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ThreeThirteenCuiController スリー・サーティーン CUI コントローラー
type ThreeThirteenCuiController struct {
	ci usecase.ThreeThirteenInteractorIF
}

// NewThreeThirteenCuiController コンストラクタ
func NewThreeThirteenCuiController(ci usecase.ThreeThirteenInteractorIF) *ThreeThirteenCuiController {
	return &ThreeThirteenCuiController{ci: ci}
}

// Exec コマンドを実行する
func (c *ThreeThirteenCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"ds", "drawstock",
			"dd", "drawdiscard",
			"d", "discard",
			"k", "knock",
			"nr", "nextround",
			"sd", "setdifficulty", "sp", "setplayers", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "ds", "drawstock":
				return c.ci.DrawFromStock(), true
			case "dd", "drawdiscard":
				return c.ci.DrawFromDiscard(), true
			case "d", "discard":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Discard)
			case "k", "knock":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Knock)
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.ThreeThirteenCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sp", "setplayers":
				return cuiutil.WithParsedInt(args, "Player count is required (2-4).", "Invalid player count: %s. Please enter 2-4.", domain.ThreeThirteenMinPlayers, domain.ThreeThirteenMaxPlayers, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.PlayerCount = v
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.ci.ActionLog)
			}
		},
	)
}
