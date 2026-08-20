//go:build !js || !wasm || extra

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KalookiCuiController カルーキ CUI コントローラー
type KalookiCuiController struct {
	ci usecase.KalookiInteractorIF
}

// NewKalookiCuiController コンストラクタ
func NewKalookiCuiController(ci usecase.KalookiInteractorIF) *KalookiCuiController {
	return &KalookiCuiController{ci: ci}
}

// Exec コマンドを実行する
func (c *KalookiCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"ds", "drawstock",
			"dd", "drawdiscard",
			"m", "meld",
			"lo", "layoff",
			"d", "discard",
			"nr", "nextround",
			"sd", "setdifficulty", "sp", "setplayers", "st", "setthreshold", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "ds", "drawstock":
				return c.ci.DrawFromStock(), true
			case "dd", "drawdiscard":
				return c.ci.DrawFromDiscard(), true
			case "m", "meld":
				groups, ok := parseSlotIndices(args)
				if !ok {
					return invalidArg("usageMABCDEFOneMeldPerArg"), true
				}
				return c.ci.Meld(groups), true
			case "lo", "layoff":
				idx := parseIntList(args)
				if len(idx) < 3 {
					return invalidArg("usageLoTargetplayeridxMeldidxCardindex"), true
				}
				return c.ci.Layoff(idx[0], idx[1], idx[2]), true
			case "d", "discard":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Discard)
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.KalookiCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sp", "setplayers":
				return cuiutil.WithParsedIntKeys(args, "playerCountRequired", "invalidPlayerCount", domain.KalookiMinPlayers, domain.KalookiMaxPlayers, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.PlayerCount = v
					return c.ci.ResetWithConfig(cfg)
				})
			case "st", "setthreshold":
				return cuiutil.WithParsedIntKeys(args, "openingThresholdRequired", "invalidThreshold0OrMore", 0, 1000, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.OpeningThreshold = v
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.ci.ActionLog)
			}
		},
	)
}
