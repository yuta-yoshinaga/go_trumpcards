//go:build !js || !wasm || extra

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GaigelCuiController ガイゲルCUIコントローラークラス
type GaigelCuiController struct {
	gi usecase.GaigelInteractorIF
}

// NewGaigelCuiController コンストラクタ
func NewGaigelCuiController(gi usecase.GaigelInteractorIF) *GaigelCuiController {
	return &GaigelCuiController{gi: gi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit           → ゲーム終了 ("bye.")
//	r / reset          → ゲームリセット (設定保持)
//	p / play <i>       → カードをプレイ
//	m / marriage <i>   → マリアージュを宣言 (リード番のみ)
//	n / next           → 次のトリックへ
//	nr / nextround     → 次のラウンドへ
//	sd / setdifficulty <0-2> → CPU難易度設定
//	st / settarget <n> → ターゲットスコア設定 (デフォルト101)
//	h / hint           → ヒント表示
//	log / l            → 棋譜表示
func (c *GaigelCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.gi.GetConfig()
			return c.gi.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play",
			"m", "marriage",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "st", "settarget",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.gi.Play)
			case "m", "marriage":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.gi.DeclareMarriage)
			case "n", "next":
				return c.gi.NextTrick(), true
			case "nr", "nextround":
				return c.gi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.gi.GetConfig()
					cfg.CpuDifficulty = domain.GaigelCpuDifficulty(v)
					return c.gi.ResetWithConfig(cfg)
				})
			case "st", "settarget":
				return cuiutil.WithParsedIntKeys(args, "targetScoreRequired", "invalidTargetScore", 1, math.MaxInt, func(v int) string {
					cfg := c.gi.GetConfig()
					cfg.TargetScore = v
					return c.gi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.gi.Hint, c.gi.ActionLog)
			}
		},
	)
}
