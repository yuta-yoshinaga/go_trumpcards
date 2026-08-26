//go:build !js || !wasm || extra2

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RistikontraCuiController はリスティコントラ CUI コントローラークラス。
type RistikontraCuiController struct {
	pi usecase.RistikontraInteractorIF
}

// NewRistikontraCuiController コンストラクタ。
func NewRistikontraCuiController(pi usecase.RistikontraInteractorIF) *RistikontraCuiController {
	return &RistikontraCuiController{pi: pi}
}

// Exec コマンド実行。
//
//	p <h>            手札 h 番を場へ出す
//	reset / r        ゲームをリセットする
//	next / n         次のゲームを開始する
//	sd <0-2>         CPU 難易度 (0=Easy, 1=Normal, 2=Hard)
//	sp <2-4>         プレイヤー数を設定する
//	log / l          棋譜を表示する
func (c *RistikontraCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.pi.GetConfig()
			return c.pi.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "next", "n", "sd", "setdifficulty", "log", "l",
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
					cfg.CpuDifficulty = domain.RistikontraCpuDifficulty(v)
					return c.pi.ResetWithConfig(cfg)
				})
			// **席数を変えるコマンドは無い。** リスティコントラは常に 2 対 2 の
			// 固定パートナーシップで、クローン元のピシュティのように 2〜4 人を
			// 選べない。宣伝だけ残すと「効かないコマンド」になる。
			default:
				return handleCuiLog(cmd, c.pi.ActionLog)
			}
		},
	)
}

// handlePlay は `p <h>` を処理する。
func (c *RistikontraCuiController) handlePlay(args []string) (string, bool) {
	if len(args) < 1 {
		return invalidArg("usagePHandidx"), true
	}
	handIdx, _, ok := cuiutil.ParseIntArgKeys([]string{args[0]}, "handIndexRequired", "invalidHandIndex", 0, 51)
	if !ok {
		return invalidArg("invalidHandIndexRaw", "val", args[0]), true
	}
	return c.pi.Play(handIdx), true
}
