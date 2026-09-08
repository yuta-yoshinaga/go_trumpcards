//go:build !js || !wasm || extra5

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OmiCuiController オミCUIコントローラークラス
type OmiCuiController struct {
	ei usecase.OmiInteractorIF
}

// NewOmiCuiController コンストラクタ
func NewOmiCuiController(ei usecase.OmiInteractorIF) *OmiCuiController {
	return &OmiCuiController{ei: ei}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit           → ゲーム終了 ("bye.")
//	r / reset          → ゲームリセット (設定保持)
//	t / trump <s>      → 切り札スートを宣言する (1-4, 1=♠ 2=♣ 3=♥ 4=♦)
//	p / play <i>       → カードをプレイ
//	n / next           → 次のトリックへ
//	nr / nextround     → 次のラウンドへ
//	sd / setdifficulty <0-2> → CPU難易度設定
//	sl / setlimit <n>  → ポイント上限設定
//	h / hint           → ヒント表示
//	log / l            → 棋譜表示
func (c *OmiCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ei.GetConfig()
			return c.ei.ResetWithConfig(cfg)
		},
		[]string{
			"t", "trump", "c", "call",
			"p", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "t", "trump", "c", "call":
				return cuiutil.WithParsedIntKeys(args, "trumpSuitRequiredRange", "invalidTrumpSuitRange", 1, 4, c.ei.CallTrump)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ei.Play)
			case "n", "next":
				return c.ei.NextTrick(), true
			case "nr", "nextround":
				return c.ei.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ei.GetConfig()
					cfg.CpuDifficulty = domain.OmiCpuDifficulty(v)
					return c.ei.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedIntKeys(args, "pointLimitRequired", "invalidPointLimit", 1, math.MaxInt, func(v int) string {
					cfg := c.ei.GetConfig()
					cfg.PointLimit = v
					return c.ei.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.ei.Hint, c.ei.ActionLog)
			}
		},
	)
}
