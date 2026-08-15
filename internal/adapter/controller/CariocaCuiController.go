//go:build !js || !wasm || extra

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CariocaCuiController カリオカ CUI コントローラー
type CariocaCuiController struct {
	ci usecase.CariocaInteractorIF
}

// NewCariocaCuiController コンストラクタ
func NewCariocaCuiController(ci usecase.CariocaInteractorIF) *CariocaCuiController {
	return &CariocaCuiController{ci: ci}
}

// Exec コマンドを実行する
func (c *CariocaCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"ds", "drawstock",
			"dd", "drawdiscard",
			"mc", "meldcontract",
			"me", "meldextra",
			"lo", "layoff",
			"d", "discard",
			"nr", "nextround",
			"pc", "setplayers", "sd", "setdifficulty", "sp", "setpenalty", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "ds", "drawstock":
				return c.ci.DrawFromStock(), true
			case "dd", "drawdiscard":
				return c.ci.DrawFromDiscard(), true
			case "mc", "meldcontract":
				slots, ok := parseSlotIndices(args)
				if !ok {
					return "Usage: mc <a,b,c> <d,e,f> [<g,h,i>]  (one slot per arg)", true
				}
				return c.ci.MeldContract(slots), true
			case "me", "meldextra":
				return c.ci.MeldExtra(parseIntList(args)), true
			case "lo", "layoff":
				idx := parseIntList(args)
				if len(idx) < 3 {
					return "Usage: lo <targetPlayerIdx> <meldIdx> <cardIndex>", true
				}
				return c.ci.Layoff(idx[0], idx[1], idx[2]), true
			case "d", "discard":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Discard)
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "pc", "setplayers":
				return cuiutil.WithParsedInt(args, "Player count is required (3-6).", "Invalid player count: %s. Please enter 3-6.", domain.CariocaPlayerCountMin, domain.CariocaPlayerCountMax, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.PlayerCount = v
					return c.ci.ResetWithConfig(cfg)
				})
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.CariocaCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sp", "setpenalty":
				return cuiutil.WithParsedInt(args, "Fail penalty is required.", "Invalid fail penalty: %s. Please enter 0 or more.", 0, 1000, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.FailContractPenalty = v
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.ci.ActionLog)
			}
		},
	)
}
