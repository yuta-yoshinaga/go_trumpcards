//go:build !js || !wasm || classic

package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ScopaCuiController スコパ CUI コントローラークラス。
type ScopaCuiController struct {
	si usecase.ScopaInteractorIF
}

// NewScopaCuiController コンストラクタ。
func NewScopaCuiController(si usecase.ScopaInteractorIF) *ScopaCuiController {
	return &ScopaCuiController{si: si}
}

// Exec コマンド実行。
//
//	play <h> [t1 t2 ...]   手札 h を出す (場札 t... を捕獲、無指定なら場に置く)
//	reset / r / next / n
//	sd <0-2>               CPU 難易度
//	log / l
func (c *ScopaCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.si.GetConfig()
			return c.si.ResetWithConfig(cfg)
		},
		[]string{
			"play", "p", "next", "n", "sd", "setdifficulty", "st", "settarget", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return c.handlePlay(args)
			case "n", "next":
				return c.si.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.si.GetConfig()
					cfg.CpuDifficulty = domain.ScopaCpuDifficulty(v)
					return c.si.ResetWithConfig(cfg)
				})
			case "st", "settarget":
				return c.setTargetScore(args), true
			default:
				return handleCuiHintAndLog(cmd, c.si.Hint, c.si.ActionLog)
			}
		},
	)
}

// setTargetScore は目標点を設定してリセットする。範囲はドメインの Validate と
// 同じ定数を読むので、片方だけ動くことがない。値は文言に埋め込まず
// {{min}}/{{max}} で渡す (#5619)。
func (c *ScopaCuiController) setTargetScore(args []string) string {
	if len(args) < 1 {
		return invalidArg("targetScoreRequired")
	}
	v, err := strconv.Atoi(args[0])
	if err != nil || v < domain.ScopaMinTargetScore || v > domain.ScopaMaxTargetScore {
		return invalidArg("invalidTargetScoreRange",
			"val", args[0],
			"min", strconv.Itoa(domain.ScopaMinTargetScore),
			"max", strconv.Itoa(domain.ScopaMaxTargetScore))
	}
	cfg := c.si.GetConfig()
	cfg.TargetScore = v
	return c.si.ResetWithConfig(cfg)
}

// handlePlay は `p <h> [t1 t2 ...]` を処理する。
func (c *ScopaCuiController) handlePlay(args []string) (string, bool) {
	if len(args) < 1 {
		return invalidArg("usagePHandidxTableidxMany"), true
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
