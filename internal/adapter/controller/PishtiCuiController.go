//go:build !js || !wasm || extra2

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PishtiCuiController は Pişti CUI コントローラークラス。
type PishtiCuiController struct {
	pi usecase.PishtiInteractorIF
}

// NewPishtiCuiController コンストラクタ。
func NewPishtiCuiController(pi usecase.PishtiInteractorIF) *PishtiCuiController {
	return &PishtiCuiController{pi: pi}
}

// Exec コマンド実行。
//
//	p <h>            手札 h 番を場へ出す
//	reset / r        ゲームをリセットする
//	next / n         次のゲームを開始する
//	sd <0-2>         CPU 難易度 (0=Easy, 1=Normal, 2=Hard)
//	sp <2-4>         プレイヤー数を設定する
//	log / l          棋譜を表示する
func (c *PishtiCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.pi.GetConfig()
			return c.pi.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "next", "n", "sd", "setdifficulty", "sp", "setplayers", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return c.handlePlay(args)
			case "n", "next":
				return c.pi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.pi.GetConfig()
					cfg.CpuDifficulty = domain.PishtiCpuDifficulty(v)
					return c.pi.ResetWithConfig(cfg)
				})
			case "sp", "setplayers":
				return cuiutil.WithParsedInt(args, "player count is required (2-4).", "Invalid player count: %s. Please enter 2-4.", domain.PishtiMinPlayerCnt, domain.PishtiMaxPlayerCnt, func(v int) string {
					cfg := c.pi.GetConfig()
					cfg.PlayerCnt = v
					return c.pi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.pi.ActionLog)
			}
		},
	)
}

// handlePlay は `p <h>` を処理する。
func (c *PishtiCuiController) handlePlay(args []string) (string, bool) {
	if len(args) < 1 {
		return "Usage: p <handIdx>", true
	}
	handIdx, _, ok := cuiutil.ParseIntArg([]string{args[0]}, "hand index is required", "Invalid hand index: %s", 0, 51)
	if !ok {
		return "Invalid hand index: " + args[0], true
	}
	return c.pi.Play(handIdx), true
}
