//go:build !js || !wasm || extra2

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CuarentaCuiController クアレンタ CUI コントローラークラス。
type CuarentaCuiController struct {
	ci usecase.CuarentaInteractorIF
}

// NewCuarentaCuiController コンストラクタ。
func NewCuarentaCuiController(ci usecase.CuarentaInteractorIF) *CuarentaCuiController {
	return &CuarentaCuiController{ci: ci}
}

// Exec コマンド実行。
//
//	play <h> / p <h>       手札 h を出す (同ランク捕獲、無ければ場に置く)
//	reset / r / next / n
//	sd <0-2>               CPU 難易度
//	log / l
func (c *CuarentaCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"play", "p", "next", "n", "sd", "setdifficulty", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return c.handlePlay(args)
			case "n", "next":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.CuarentaCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.ci.ActionLog)
			}
		},
	)
}

// handlePlay は `p <h>` を処理する。
func (c *CuarentaCuiController) handlePlay(args []string) (string, bool) {
	if len(args) < 1 {
		return invalidArg("usagePHandidx"), true
	}
	handIdx, _, ok := cuiutil.ParseIntArgKeys([]string{args[0]}, "handIndexRequired", "invalidHandIndex", 0, 39)
	if !ok {
		return invalidArg("invalidHandIndexRaw", "val", args[0]), true
	}
	return c.ci.Play(handIdx), true
}
