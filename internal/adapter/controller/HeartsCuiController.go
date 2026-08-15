package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HeartsCuiController ハーツCUIコントローラークラス
type HeartsCuiController struct {
	hi usecase.HeartsInteractorIF
}

// NewHeartsCuiController コンストラクタ
func NewHeartsCuiController(hi usecase.HeartsInteractorIF) *HeartsCuiController {
	return &HeartsCuiController{hi: hi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit         → ゲーム終了 ("bye.")
//	r / reset        → ゲームリセット (設定保持)
//	pass <i1> <i2> <i3> → カード3枚を交換
//	p / play <i>     → カードをプレイ
//	n / next         → 次のトリックへ
//	nr / nextround   → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	sl / setlimit <n> → ポイント上限設定
//	h / hint         → ヒント表示
//	log / l          → 棋譜表示
func (c *HeartsCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.hi.GetConfig()
			return c.hi.ResetWithConfig(cfg)
		},
		[]string{
			"pass", "p", "play", "n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "pass":
				indices, skipped := cuiutil.ParseIntSlice(args)
				var result string
				if len(indices) != 3 {
					result = "Pass requires exactly 3 card indices."
				} else {
					result = c.hi.Pass(indices)
				}
				return cuiutil.PrependSkippedWarning(result, skipped), true
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.hi.Play)
			case "n", "next":
				return c.hi.NextTrick(), true
			case "nr", "nextround":
				return c.hi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.hi.GetConfig()
					cfg.CpuDifficulty = domain.HeartsCpuDifficulty(v)
					return c.hi.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedInt(args, "Point limit is required.", "Invalid point limit: %s. Please enter 1 or more.", 1, math.MaxInt, func(v int) string {
					cfg := c.hi.GetConfig()
					cfg.PointLimit = v
					return c.hi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.hi.Hint, c.hi.ActionLog)
			}
		},
	)
}
