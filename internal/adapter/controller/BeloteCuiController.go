//go:build !js || !wasm || extra3

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BeloteCuiController ベロートCUIコントローラークラス
type BeloteCuiController struct {
	bi usecase.BeloteInteractorIF
}

// NewBeloteCuiController コンストラクタ
func NewBeloteCuiController(bi usecase.BeloteInteractorIF) *BeloteCuiController {
	return &BeloteCuiController{bi: bi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit           → ゲーム終了 ("bye.")
//	r / reset          → ゲームリセット (設定保持)
//	o / orderup        → ピックアップ (ターンアップのスートを切り札に指名)
//	pa / pass          → パス (フェーズに応じて PickUp(false) or PassCall)
//	c / call <suit>    → スートをコール (1=Spade, 2=Club, 3=Heart, 4=Diamond)
//	p / play <i>       → カードをプレイ
//	n / next           → 次のトリックへ
//	nr / nextround     → 次のラウンドへ
//	sd / setdifficulty <0-2> → CPU難易度設定
//	st / settarget <n> → ターゲットスコア設定 (デフォルト1000)
//	h / hint           → ヒント表示
//	log / l            → 棋譜表示
func (c *BeloteCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.bi.GetConfig()
			return c.bi.ResetWithConfig(cfg)
		},
		[]string{
			"o", "orderup",
			"pa", "pass",
			"c", "call",
			"p", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "st", "settarget",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "o", "orderup":
				return c.bi.PickUp(true), true
			case "pa", "pass":
				return c.bi.Pass(), true
			case "c", "call":
				return cuiutil.WithParsedInt(args, "Suit is required (1-4).", "Invalid suit: %s.", 1, 4, func(suit int) string {
					return c.bi.CallTrump(suit)
				})
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.bi.Play)
			case "n", "next":
				return c.bi.NextTrick(), true
			case "nr", "nextround":
				return c.bi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.bi.GetConfig()
					cfg.CpuDifficulty = domain.BeloteCpuDifficulty(v)
					return c.bi.ResetWithConfig(cfg)
				})
			case "st", "settarget":
				return cuiutil.WithParsedInt(args, "Target score is required.", "Invalid target score: %s. Please enter 1 or more.", 1, math.MaxInt, func(v int) string {
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
