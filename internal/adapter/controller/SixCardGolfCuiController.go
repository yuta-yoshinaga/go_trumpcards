package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SixCardGolfCuiController SixCardGolf CUIコントローラー
type SixCardGolfCuiController struct {
	ci usecase.SixCardGolfInteractorIF
}

// NewSixCardGolfCuiController コンストラクタ
func NewSixCardGolfCuiController(ci usecase.SixCardGolfInteractorIF) *SixCardGolfCuiController {
	return &SixCardGolfCuiController{ci: ci}
}

// Exec コマンド実行
func (c *SixCardGolfCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"fi", "flipinitial", "ds", "drawstock", "dd", "drawdiscard",
			"sw", "swap", "di", "discard", "fl", "flip", "sf", "skipflip",
			"nr", "nextround",
			"sd", "setdifficulty", "sp", "setplayers", "sr", "setrounds",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "fi", "flipinitial":
				return cuiutil.WithParsedInt(args, "Position is required (0-5).", "Invalid position: %s.", 0, 5, c.ci.FlipInitial)
			case "ds", "drawstock":
				return c.ci.DrawStock(), true
			case "dd", "drawdiscard":
				return c.ci.DrawDiscard(), true
			case "sw", "swap":
				return cuiutil.WithParsedInt(args, "Position is required (0-5).", "Invalid position: %s.", 0, 5, c.ci.SwapCard)
			case "di", "discard":
				return c.ci.DiscardDrawn(), true
			case "fl", "flip":
				return cuiutil.WithParsedInt(args, "Position is required (0-5).", "Invalid position: %s.", 0, 5, c.ci.FlipCard)
			case "sf", "skipflip":
				return c.ci.SkipFlip(), true
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.SixCardGolfCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sp", "setplayers":
				return cuiutil.WithParsedInt(args, "Player count is required (2-4).", "Invalid player count: %s. Please enter 2-4.", 2, 4, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.PlayerCount = v
					return c.ci.ResetWithConfig(cfg)
				})
			case "sr", "setrounds":
				return cuiutil.WithParsedInt(args, "Round count is required (1-18).", "Invalid round count: %s. Please enter 1-18.", 1, 18, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.Rounds = v
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.ci.Hint, c.ci.ActionLog)
			}
		},
	)
}
