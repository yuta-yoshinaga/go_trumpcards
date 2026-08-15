//go:build !js || !wasm || classic

package controller

import (
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ScoponeCuiController スコポーネ CUI コントローラークラス。
type ScoponeCuiController struct {
	si usecase.ScoponeInteractorIF
}

// NewScoponeCuiController コンストラクタ。
func NewScoponeCuiController(si usecase.ScoponeInteractorIF) *ScoponeCuiController {
	return &ScoponeCuiController{si: si}
}

// Exec コマンド実行。
//
//	play <h> [t1 t2 ...]   手札 h を出す (場札 t... を捕獲、無指定なら場に置く)
//	reset / r / next / n / nextround
//	sd <0-2>               CPU 難易度
//	st <n>                 目標スコア
//	log / l
func (c *ScoponeCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.si.GetConfig()
			return c.si.ResetWithConfig(cfg)
		},
		[]string{
			"play", "p", "next", "n", "nextround",
			"sd", "setdifficulty", "st", "settarget", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return c.handlePlay(args)
			case "n", "next", "nextround":
				return c.si.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.si.GetConfig()
					cfg.CpuDifficulty = domain.ScoponeCpuDifficulty(v)
					return c.si.ResetWithConfig(cfg)
				})
			case "st", "settarget":
				return cuiutil.WithParsedIntKeys(args, "targetScoreRequiredAlt", "invalidTargetScorePlain", 1, domain.ScoponeMaxTargetScore, func(v int) string {
					cfg := c.si.GetConfig()
					cfg.TargetScore = v
					return c.si.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.si.Hint, c.si.ActionLog)
			}
		},
	)
}

// handlePlay は `p <h> [t1 t2 ...]` を処理する。
func (c *ScoponeCuiController) handlePlay(args []string) (string, bool) {
	if len(args) < 1 {
		return "Usage: p <handIdx> [tableIdx...]", true
	}
	handIdx, _, ok := cuiutil.ParseIntArgKeys([]string{args[0]}, "handIndexRequired", "invalidHandIndex", 0, 39)
	if !ok {
		return invalidArg("invalidHandIndexRaw", "val", args[0]), true
	}
	tableIdxs, skipped := cuiutil.ParseIntSlice(args[1:])
	// Refuse before playing. PrependSkippedWarning ran the move first and
	// put the warning above the new board, so a mistyped index was dropped
	// and the remaining ones played as a different, legal move (issue #5390).
	if len(skipped) > 0 {
		return invalidArg("invalidCardIndex", "val", strings.Join(skipped, ", ")), true
	}
	return c.si.Play(handIdx, tableIdxs), true
}
