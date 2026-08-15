//go:build !js || !wasm || extra3

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CribbageCuiController クリベッジCUIコントローラークラス
type CribbageCuiController struct {
	ci usecase.CribbageInteractorIF
}

// NewCribbageCuiController コンストラクタ
func NewCribbageCuiController(ci usecase.CribbageInteractorIF) *CribbageCuiController {
	return &CribbageCuiController{ci: ci}
}

// Exec コマンド実行
func (c *CribbageCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"d", "discard", "c", "cut", "p", "peg", "go",
			"sn", "shownext",
			"nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit", "log", "l", "h", "hint",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "discard":
				indices := parseIntList(args)
				return c.ci.Discard(indices), true
			case "c", "cut":
				return c.ci.Cut(), true
			case "p", "peg":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Peg)
			case "go":
				return c.ci.Go(), true
			case "h", "hint":
				return c.ci.Hint(), true
			case "sn", "shownext":
				return c.ci.ShowNext(), true
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.CribbageCpuDifficulty(v)
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
