//go:build !js || !wasm || extra3

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// JassCuiController ヤス(シーバー)CUIコントローラークラス
type JassCuiController struct {
	ji usecase.JassInteractorIF
}

// NewJassCuiController コンストラクタ
func NewJassCuiController(ji usecase.JassInteractorIF) *JassCuiController {
	return &JassCuiController{ji: ji}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit           → ゲーム終了 ("bye.")
//	r / reset          → ゲームリセット (設定保持)
//	c / call <suit>    → 切り札スートを指名 (1=Spade, 2=Club, 3=Heart, 4=Diamond)
//	sc / schieben      → パートナーへ切り札選択を委譲 (フォアハンドのみ)
//	p / play <i>       → カードをプレイ
//	n / next           → 次のトリックへ
//	nr / nextround     → 次のラウンドへ
//	sd / setdifficulty <0-2> → CPU難易度設定
//	st / settarget <n> → ターゲットスコア設定 (デフォルト1000)
//	h / hint           → ヒント表示
//	log / l            → 棋譜表示
func (c *JassCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ji.GetConfig()
			return c.ji.ResetWithConfig(cfg)
		},
		[]string{
			"c", "call",
			"sc", "schieben",
			"p", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "st", "settarget",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "c", "call":
				return cuiutil.WithParsedInt(args, "Suit is required (1-4).", "Invalid suit: %s.", 1, 4, func(suit int) string {
					return c.ji.ChooseTrump(suit)
				})
			case "sc", "schieben":
				return c.ji.Schieben(), true
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ji.Play)
			case "n", "next":
				return c.ji.NextTrick(), true
			case "nr", "nextround":
				return c.ji.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ji.GetConfig()
					cfg.CpuDifficulty = domain.JassCpuDifficulty(v)
					return c.ji.ResetWithConfig(cfg)
				})
			case "st", "settarget":
				return cuiutil.WithParsedIntKeys(args, "targetScoreRequired", "invalidTargetScore", 1, math.MaxInt, func(v int) string {
					cfg := c.ji.GetConfig()
					cfg.TargetScore = v
					return c.ji.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.ji.Hint, c.ji.ActionLog)
			}
		},
	)
}
