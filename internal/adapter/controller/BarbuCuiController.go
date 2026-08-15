//go:build !js || !wasm || solo

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BarbuCuiController はバルブ CUI コントローラークラス。
type BarbuCuiController struct {
	bi usecase.BarbuInteractorIF
}

// NewBarbuCuiController コンストラクタ。
func NewBarbuCuiController(bi usecase.BarbuInteractorIF) *BarbuCuiController {
	return &BarbuCuiController{bi: bi}
}

// Exec コマンド実行。
//
//	contract <c> [trump]  コントラクト c を選択 (Trumps=5 のみ trump 1-4 を指定)
//	play <h>              手札 h を出す (Dominoes では -1 でパス)
//	next / n              次のディール
//	sd <0-2>              CPU 難易度
//	log / l
func (c *BarbuCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.bi.GetConfig()
			return c.bi.ResetWithConfig(cfg)
		},
		[]string{
			"contract", "c", "play", "p", "next", "n", "sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "c", "contract":
				return c.handleSelectContract(args)
			case "p", "play":
				return c.handlePlay(args)
			case "n", "next":
				return c.bi.NextDeal(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.bi.GetConfig()
					cfg.CpuDifficulty = domain.BarbuCpuDifficulty(v)
					return c.bi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.bi.Hint, c.bi.ActionLog)
			}
		},
	)
}

// handleSelectContract は `c <contract> [trump]` を処理する。
func (c *BarbuCuiController) handleSelectContract(args []string) (string, bool) {
	if len(args) < 1 {
		return "Usage: c <contract 0-6> [trumpSuit 1-4 for Trumps]", true
	}
	contract, _, ok := cuiutil.ParseIntArgKeys([]string{args[0]}, "contractRequired", "invalidContract", 0, domain.BarbuContractCnt-1)
	if !ok {
		return invalidArg("invalidContractRaw", "val", args[0]), true
	}
	trump := -1
	if len(args) >= 2 {
		t, _, tok := cuiutil.ParseIntArgKeys([]string{args[1]}, "", "invalidTrumpSuitNoPeriod", domain.CardDesignSpade, domain.CardDesignDiamond)
		if tok {
			trump = t
		}
	}
	return c.bi.SelectContract(contract, trump), true
}

// handlePlay は `p <h>` (Dominoes では `p -1` でパス) を処理する。
func (c *BarbuCuiController) handlePlay(args []string) (string, bool) {
	if len(args) < 1 {
		return "Usage: p <handIdx> (-1 to pass in Dominoes)", true
	}
	handIdx, _, ok := cuiutil.ParseIntArgKeys([]string{args[0]}, "handIndexRequired", "invalidHandIndex", -1, domain.BarbuHandSize-1)
	if !ok {
		return invalidArg("invalidHandIndexRaw", "val", args[0]), true
	}
	return c.bi.Play(handIdx, nil), true
}
