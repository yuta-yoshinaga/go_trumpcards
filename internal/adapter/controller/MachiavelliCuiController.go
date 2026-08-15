//go:build !js || !wasm || extra

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MachiavelliCuiController マキャヴェッリ CUI コントローラー
type MachiavelliCuiController struct {
	ci usecase.MachiavelliInteractorIF
}

// NewMachiavelliCuiController コンストラクタ
func NewMachiavelliCuiController(ci usecase.MachiavelliInteractorIF) *MachiavelliCuiController {
	return &MachiavelliCuiController{ci: ci}
}

// Exec コマンドを実行する
func (c *MachiavelliCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"dr", "draw",
			"nm", "newmeld",
			"lo", "layoff",
			"nr", "nextround",
			"pc", "setplayers", "sd", "setdifficulty", "sr", "setrounds", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "dr", "draw":
				return c.ci.Draw(), true
			case "nm", "newmeld":
				idx := parseIntList(args)
				if len(idx) < domain.MachiavelliMeldMin {
					return invalidArg("usageNmIJKAtLeast3HandIndices"), true
				}
				return c.ci.NewMeld(idx), true
			case "lo", "layoff":
				idx := parseIntList(args)
				if len(idx) < 2 {
					return invalidArg("usageLoMeldidxHandindex"), true
				}
				return c.ci.Layoff(idx[0], idx[1]), true
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "pc", "setplayers":
				return cuiutil.WithParsedIntKeys(args, "playerCountRequired25", "invalidPlayerCount25", domain.MachiavelliPlayerCountMin, domain.MachiavelliPlayerCountMax, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.PlayerCount = v
					return c.ci.ResetWithConfig(cfg)
				})
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.MachiavelliCpuDifficulty(v)
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
