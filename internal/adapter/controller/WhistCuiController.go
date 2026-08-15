package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// WhistCuiController ホイストCUIコントローラークラス
type WhistCuiController struct {
	wi usecase.WhistInteractorIF
}

// NewWhistCuiController コンストラクタ
func NewWhistCuiController(wi usecase.WhistInteractorIF) *WhistCuiController {
	return &WhistCuiController{wi: wi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit         → ゲーム終了 ("bye.")
//	r / reset        → ゲームリセット (設定保持)
//	p / play <i>     → カードをプレイ
//	n / next         → 次のトリックへ
//	nr / nextround   → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	sl / setlimit <n> → ポイント上限設定
//	h / hint         → ヒント表示
//	log / l          → 棋譜表示
func (c *WhistCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.wi.GetConfig()
			return c.wi.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.wi.Play)
			case "n", "next":
				return c.wi.NextTrick(), true
			case "nr", "nextround":
				return c.wi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.wi.GetConfig()
					cfg.CpuDifficulty = domain.WhistCpuDifficulty(v)
					return c.wi.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedInt(args, "Point limit is required.", "Invalid point limit: %s. Please enter 1 or more.", 1, math.MaxInt, func(v int) string {
					cfg := c.wi.GetConfig()
					cfg.PointLimit = v
					return c.wi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.wi.Hint, c.wi.ActionLog)
			}
		},
	)
}
