//go:build !js || !wasm || extra

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// IndianRummyCuiController インドラミー CUI コントローラー
type IndianRummyCuiController struct {
	ci usecase.IndianRummyInteractorIF
}

// NewIndianRummyCuiController コンストラクタ
func NewIndianRummyCuiController(ci usecase.IndianRummyInteractorIF) *IndianRummyCuiController {
	return &IndianRummyCuiController{ci: ci}
}

// Exec コマンドを実行する
func (c *IndianRummyCuiController) Exec(command string) string {
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
			"de", "declare",
			"nr", "nextround",
			"pc", "setplayers", "sd", "setdifficulty", "sr", "setrounds", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "ds", "drawstock":
				return c.ci.DrawFromStock(), true
			case "dd", "drawdiscard":
				return c.ci.DrawFromDiscard(), true
			case "d", "discard":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Discard)
			case "de", "declare":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Declare)
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "pc", "setplayers":
				return cuiutil.WithParsedInt(args, "Player count is required (2-4).", "Invalid player count: %s. Please enter 2-4.", domain.IndianRummyPlayerCountMin, domain.IndianRummyPlayerCountMax, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.PlayerCount = v
					return c.ci.ResetWithConfig(cfg)
				})
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.IndianRummyCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sr", "setrounds":
				return cuiutil.WithParsedInt(args, "Target rounds is required.", "Invalid target rounds: %s. Please enter 1 or more.", 1, math.MaxInt, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.TargetRounds = v
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.ci.ActionLog)
			}
		},
	)
}
