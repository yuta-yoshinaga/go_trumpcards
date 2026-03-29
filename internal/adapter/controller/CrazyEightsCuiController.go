package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CrazyEightsCuiController クレイジーエイトCUIコントローラークラス
type CrazyEightsCuiController struct {
	ci usecase.CrazyEightsInteractorIF
}

// NewCrazyEightsCuiController コンストラクタ
func NewCrazyEightsCuiController(ci usecase.CrazyEightsInteractorIF) *CrazyEightsCuiController {
	return &CrazyEightsCuiController{ci: ci}
}

// Exec コマンド実行
func (c *CrazyEightsCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "d", "draw", "s", "suit",
			"nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax, c.ci.Play)
			case "d", "draw":
				return c.ci.Draw(), true
			case "s", "suit":
				return cuiutil.WithParsedInt(args, "Suit is required (1=♠, 2=♣, 3=♥, 4=♦).", "Invalid suit: %s. Please enter 1-4.", 1, 4, c.ci.ChooseSuit)
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.CrazyEightsCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedInt(args, "Point limit is required.", "Invalid point limit: %s. Please enter 1 or more.", 1, math.MaxInt, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.PointLimit = v
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.ci.ActionLog)
			}
		},
	)
}
