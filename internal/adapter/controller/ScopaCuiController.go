//go:build !js || !wasm || classic

package controller

import (
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
			"play", "p", "next", "n", "sd", "setdifficulty", "h", "hint", "log", "l",
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
			default:
				return handleCuiHintAndLog(cmd, c.si.Hint, c.si.ActionLog)
			}
		},
	)
}

// handlePlay は `p <h> [t1 t2 ...]` を処理する。
func (c *ScopaCuiController) handlePlay(args []string) (string, bool) {
	if len(args) < 1 {
		return "Usage: p <handIdx> [tableIdx...]", true
	}
	handIdx, _, ok := cuiutil.ParseIntArg([]string{args[0]}, "hand index is required", "Invalid hand index: %s", 0, 39)
	if !ok {
		return "Invalid hand index: " + args[0], true
	}
	tableIdxs, skipped := cuiutil.ParseIntSlice(args[1:])
	return cuiutil.PrependSkippedWarning(c.si.Play(handIdx, tableIdxs), skipped), true
}
