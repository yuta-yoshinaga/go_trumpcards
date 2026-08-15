//go:build !js || !wasm || classic

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// EscobaCuiController エスコバ CUI コントローラークラス。
type EscobaCuiController struct {
	ei usecase.EscobaInteractorIF
}

// NewEscobaCuiController コンストラクタ。
func NewEscobaCuiController(ei usecase.EscobaInteractorIF) *EscobaCuiController {
	return &EscobaCuiController{ei: ei}
}

// Exec コマンド実行。
//
//	play <h> [t1 t2 ...]   手札 h を出す (場札 t... を捕獲、無指定なら場に置く)
//	reset / r / next / n / nextround
//	sd <0-2>               CPU 難易度
//	st <n>                 目標スコア
//	log / l
func (c *EscobaCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ei.GetConfig()
			return c.ei.ResetWithConfig(cfg)
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
				return c.ei.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ei.GetConfig()
					cfg.CpuDifficulty = domain.EscobaCpuDifficulty(v)
					return c.ei.ResetWithConfig(cfg)
				})
			case "st", "settarget":
				return cuiutil.WithParsedInt(args, "target score is required.", "Invalid target score: %s.", 1, domain.EscobaMaxTargetScore, func(v int) string {
					cfg := c.ei.GetConfig()
					cfg.TargetScore = v
					return c.ei.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.ei.Hint, c.ei.ActionLog)
			}
		},
	)
}

// handlePlay は `p <h> [t1 t2 ...]` を処理する。
func (c *EscobaCuiController) handlePlay(args []string) (string, bool) {
	if len(args) < 1 {
		return "Usage: p <handIdx> [tableIdx...]", true
	}
	handIdx, _, ok := cuiutil.ParseIntArg([]string{args[0]}, "hand index is required", "Invalid hand index: %s", 0, 39)
	if !ok {
		return "Invalid hand index: " + args[0], true
	}
	tableIdxs, skipped := cuiutil.ParseIntSlice(args[1:])
	return cuiutil.PrependSkippedWarning(c.ei.Play(handIdx, tableIdxs), skipped), true
}
