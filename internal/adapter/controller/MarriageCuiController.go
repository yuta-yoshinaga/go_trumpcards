//go:build !js || !wasm || extra5

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MarriageCuiController マリッジ CUI コントローラー
type MarriageCuiController struct {
	ci usecase.MarriageInteractorIF
}

// NewMarriageCuiController コンストラクタ
func NewMarriageCuiController(ci usecase.MarriageInteractorIF) *MarriageCuiController {
	return &MarriageCuiController{ci: ci}
}

// Exec コマンドを実行する
func (c *MarriageCuiController) Exec(command string) string {
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
				return cuiutil.WithParsedIntKeys(args, "playerCountRequired", "invalidPlayerCount", domain.MarriagePlayerCountMin, domain.MarriagePlayerCountMax, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.PlayerCount = v
					return c.ci.ResetWithConfig(cfg)
				})
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.MarriageCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sr", "setrounds":
				return cuiutil.WithParsedIntKeys(args, "targetRoundsRequiredPlain", "invalidTargetRounds1OrMore", 1, math.MaxInt, func(v int) string {
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
