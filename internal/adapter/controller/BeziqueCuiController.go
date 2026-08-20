//go:build !js || !wasm || classic

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BeziqueCuiController ベジークCUIコントローラークラス
type BeziqueCuiController struct {
	bi usecase.BeziqueInteractorIF
}

// NewBeziqueCuiController コンストラクタ
func NewBeziqueCuiController(bi usecase.BeziqueInteractorIF) *BeziqueCuiController {
	return &BeziqueCuiController{bi: bi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	p / play <i>             → カードをプレイ
//	m / meld <i>             → 役をインデックス指定で宣言
//	s / skip                 → 役宣言をパス
//	n / next / nextround     → 次のディールへ
//	sd / setdifficulty <0-2> → CPU難易度設定
//	st / settarget <n>       → ターゲットスコア設定 (デフォルト1000)
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *BeziqueCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.bi.GetConfig()
			return c.bi.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "m", "meld", "s", "skip",
			"n", "next", "nextround",
			"sd", "setdifficulty", "st", "settarget",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.bi.Play)
			case "m", "meld":
				return cuiutil.WithParsedIntKeys(args, "meldIndexRequired", "invalidMeldIndex", cuiutil.NoMin, cuiutil.NoMax, c.bi.DeclareMeld)
			case "s", "skip":
				return c.bi.SkipMeld(), true
			case "n", "next", "nextround":
				return c.bi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.bi.GetConfig()
					cfg.CpuDifficulty = domain.BeziqueCpuDifficulty(v)
					return c.bi.ResetWithConfig(cfg)
				})
			case "st", "settarget":
				return cuiutil.WithParsedIntKeys(args, "targetScoreRequired", "invalidTargetScore", 1, math.MaxInt, func(v int) string {
					cfg := c.bi.GetConfig()
					cfg.TargetScore = v
					return c.bi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.bi.Hint, c.bi.ActionLog)
			}
		},
	)
}
