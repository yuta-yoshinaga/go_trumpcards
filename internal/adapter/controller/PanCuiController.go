//go:build !js || !wasm || extra

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PanCuiController パングインゲ CUI コントローラー
type PanCuiController struct {
	ci usecase.PanInteractorIF
}

// NewPanCuiController コンストラクタ
func NewPanCuiController(ci usecase.PanInteractorIF) *PanCuiController {
	return &PanCuiController{ci: ci}
}

// Exec コマンドを実行する
func (c *PanCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"ds", "drawstock", "dd", "drawdiscard",
			"m", "meld", "lo", "layoff", "d", "discard",
			"nr", "nextround",
			"pc", "setplayers", "sd", "setdifficulty", "sr", "setrounds", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "ds", "drawstock":
				return c.ci.DrawFromStock(), true
			case "dd", "drawdiscard":
				return c.ci.DrawFromDiscard(), true
			case "m", "meld":
				return c.ci.Meld(parseIntList(args)), true
			case "lo", "layoff":
				indices := parseIntList(args)
				if len(indices) != 3 {
					return "param error: layoff requires 3 ints: <owner> <meldIdx> <cardIdx>.", true
				}
				return c.ci.Layoff(indices[0], indices[1], indices[2]), true
			case "d", "discard":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Discard)
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "pc", "setplayers":
				return cuiutil.WithParsedInt(args, "Player count is required (3-6).", "Invalid player count: %s. Please enter 3-6.", domain.PanPlayerCountMin, domain.PanPlayerCountMax, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.PlayerCount = v
					return c.ci.ResetWithConfig(cfg)
				})
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.PanCpuDifficulty(v)
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
