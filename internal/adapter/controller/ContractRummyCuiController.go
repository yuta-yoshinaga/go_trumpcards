//go:build !js || !wasm || extra

package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ContractRummyCuiController コントラクトラミー CUI コントローラー
type ContractRummyCuiController struct {
	ci usecase.ContractRummyInteractorIF
}

// NewContractRummyCuiController コンストラクタ
func NewContractRummyCuiController(ci usecase.ContractRummyInteractorIF) *ContractRummyCuiController {
	return &ContractRummyCuiController{ci: ci}
}

// Exec コマンドを実行する
func (c *ContractRummyCuiController) Exec(command string) string {
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
			"sd", "setdifficulty", "sp", "setpenalty", "log", "l",
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
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.ContractRummyCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sp", "setpenalty":
				return cuiutil.WithParsedIntKeys(args, "failPenaltyRequired", "invalidFailPenalty0OrMore", 0, 1000, func(v int) string {
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

// parseSlotIndices parses slot arguments where each arg is a comma-separated index list.
// Example: ["0,1,2", "3,4,5"] → [[0,1,2], [3,4,5]]
func parseSlotIndices(args []string) ([][]int, bool) {
	if len(args) == 0 {
		return nil, false
	}
	slots := make([][]int, 0, len(args))
	for _, a := range args {
		parts := strings.Split(a, ",")
		idx := make([]int, 0, len(parts))
		for _, p := range parts {
			n, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil {
				return nil, false
			}
			idx = append(idx, n)
		}
		slots = append(slots, idx)
	}
	return slots, true
}
