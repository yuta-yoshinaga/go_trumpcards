package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// AllFoursCuiController All Fours CUIコントローラークラス
type AllFoursCuiController struct {
	ai usecase.AllFoursInteractorIF
}

// NewAllFoursCuiController コンストラクタ
func NewAllFoursCuiController(ai usecase.AllFoursInteractorIF) *AllFoursCuiController {
	return &AllFoursCuiController{ai: ai}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	stand / st               → スタンド (非親)
//	beg / bg                 → ベグ (非親)
//	gift / g                 → ギフト=1点を渡す (親の応答)
//	run                      → ランザカード=引き直す (親の応答)
//	p / play <i>             → カードをプレイ
//	n / next                 → 次のトリックへ
//	nr / nextround           → 次のディールへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	sl / setlimit <n>        → ポイント上限設定
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *AllFoursCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ai.GetConfig()
			return c.ai.ResetWithConfig(cfg)
		},
		[]string{
			"stand", "st", "beg", "bg", "gift", "g", "run",
			"p", "play", "n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "stand", "st":
				return c.ai.Beg(false), true
			case "beg", "bg":
				return c.ai.Beg(true), true
			case "gift", "g":
				return c.ai.RespondBeg(false), true
			case "run":
				return c.ai.RespondBeg(true), true
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ai.Play)
			case "n", "next":
				return c.ai.NextTrick(), true
			case "nr", "nextround":
				return c.ai.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ai.GetConfig()
					cfg.CpuDifficulty = domain.AllFoursCpuDifficulty(v)
					return c.ai.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedIntKeys(args, "pointLimitRequired", "invalidPointLimit", 1, math.MaxInt, func(v int) string {
					cfg := c.ai.GetConfig()
					cfg.PointLimit = v
					return c.ai.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.ai.Hint, c.ai.ActionLog)
			}
		},
	)
}
