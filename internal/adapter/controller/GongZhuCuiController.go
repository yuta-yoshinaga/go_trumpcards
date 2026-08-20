package controller

import (
	"math"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GongZhuCuiController 拱猪CUIコントローラークラス
type GongZhuCuiController struct {
	gi usecase.GongZhuInteractorIF
}

// NewGongZhuCuiController コンストラクタ
func NewGongZhuCuiController(gi usecase.GongZhuInteractorIF) *GongZhuCuiController {
	return &GongZhuCuiController{gi: gi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit              → ゲーム終了 ("bye.")
//	r / reset             → ゲームリセット (設定保持)
//	expose [i...]         → ポイントカードを公開 (引数なしで公開なし)
//	p / play <i>          → カードをプレイ
//	n / next              → 次のトリックへ
//	nr / nextround        → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	sl / setlimit <n>     → ポイント上限設定
//	h / hint              → ヒント表示
//	log / l               → 棋譜表示
func (c *GongZhuCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.gi.GetConfig()
			return c.gi.ResetWithConfig(cfg)
		},
		[]string{
			"expose", "p", "play", "n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "expose":
				indices, skipped := cuiutil.ParseIntSlice(args)
				// Refuse before playing. PrependSkippedWarning ran the move first and
				// put the warning above the new board, so a mistyped index was dropped
				// and the remaining ones played as a different, legal move (issue #5390).
				if len(skipped) > 0 {
					return invalidArg("invalidCardIndex", "val", strings.Join(skipped, ", ")), true
				}
				return c.gi.Expose(indices), true
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.gi.Play)
			case "n", "next":
				return c.gi.NextTrick(), true
			case "nr", "nextround":
				return c.gi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.gi.GetConfig()
					cfg.CpuDifficulty = domain.GongZhuCpuDifficulty(v)
					return c.gi.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedIntKeys(args, "pointLimitRequired", "invalidPointLimit", 1, math.MaxInt, func(v int) string {
					cfg := c.gi.GetConfig()
					cfg.PointLimit = v
					return c.gi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.gi.Hint, c.gi.ActionLog)
			}
		},
	)
}
