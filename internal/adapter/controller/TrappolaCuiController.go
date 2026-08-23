//go:build !js || !wasm || extra2

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TrappolaCuiController トラッポラのCUIコントローラークラス
type TrappolaCuiController struct {
	ti usecase.TrappolaInteractorIF
}

// NewTrappolaCuiController コンストラクタ
func NewTrappolaCuiController(ti usecase.TrappolaInteractorIF) *TrappolaCuiController {
	return &TrappolaCuiController{ti: ti}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit              → ゲーム終了 ("bye.")
//	r / reset             → ゲームリセット (設定保持)
//	p / play <i>          → カードをプレイ
//	n / next              → 次のトリックへ
//	nr / nextround        → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	st / settarget <n>    → 目標点設定
//	h / hint              → ヒント表示
//	log / l               → 棋譜表示
func (c *TrappolaCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ti.GetConfig()
			return c.ti.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "n", "next", "nr", "nextround",
			"sd", "setdifficulty", "st", "settarget", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ti.Play)
			case "n", "next":
				return c.ti.NextTrick(), true
			case "nr", "nextround":
				return c.ti.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.CpuDifficulty = domain.TrappolaCpuDifficulty(v)
					return c.ti.ResetWithConfig(cfg)
				})
			case "st", "settarget":
				return cuiutil.WithParsedIntKeys(args, "targetPointsRequired", "invalidTargetPoints1OrMore", 1, math.MaxInt, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.TargetPoints = v
					return c.ti.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.ti.Hint, c.ti.ActionLog)
			}
		},
	)
}
