//go:build !js || !wasm || extra

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KingCuiController はキング CUI コントローラークラス。
type KingCuiController struct {
	ki usecase.KingInteractorIF
}

// NewKingCuiController コンストラクタ。
func NewKingCuiController(ki usecase.KingInteractorIF) *KingCuiController {
	return &KingCuiController{ki: ki}
}

// Exec コマンド実行。
//
//	contract <c> [trump]  コントラクト c を選択 (King Trump=6 のみ trump 1-4 を指定)
//	play <h>              手札 h を出す
//	next / n              次のディール
//	sd <0-2>              CPU 難易度
//	h / hint              ヒント表示
//	log / l               棋譜表示
func (c *KingCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ki.GetConfig()
			return c.ki.ResetWithConfig(cfg)
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
				return c.ki.NextDeal(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ki.GetConfig()
					cfg.CpuDifficulty = domain.KingCpuDifficulty(v)
					return c.ki.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.ki.Hint, c.ki.ActionLog)
			}
		},
	)
}

// handleSelectContract は `c <contract> [trump]` を処理する。
func (c *KingCuiController) handleSelectContract(args []string) (string, bool) {
	if len(args) < 1 {
		return invalidArg("usageCContract06Trumpsuit14ForKingTrump"), true
	}
	contract, _, ok := cuiutil.ParseIntArgKeys([]string{args[0]}, "contractRequired", "invalidContract", 0, domain.KingContractCnt-1)
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
	return c.ki.SelectContract(contract, trump), true
}

// handlePlay は `p <h>` を処理する。
func (c *KingCuiController) handlePlay(args []string) (string, bool) {
	if len(args) < 1 {
		return invalidArg("usagePHandidx"), true
	}
	handIdx, _, ok := cuiutil.ParseIntArgKeys([]string{args[0]}, "handIndexRequired", "invalidHandIndex", 0, domain.KingHandSize-1)
	if !ok {
		return invalidArg("invalidHandIndexRaw", "val", args[0]), true
	}
	return c.ki.Play(handIdx), true
}
