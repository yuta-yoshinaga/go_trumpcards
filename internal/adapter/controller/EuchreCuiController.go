//go:build !js || !wasm || solo

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// EuchreCuiController ユーカーCUIコントローラークラス
type EuchreCuiController struct {
	ei usecase.EuchreInteractorIF
}

// NewEuchreCuiController コンストラクタ
func NewEuchreCuiController(ei usecase.EuchreInteractorIF) *EuchreCuiController {
	return &EuchreCuiController{ei: ei}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit           → ゲーム終了 ("bye.")
//	r / reset          → ゲームリセット (設定保持)
//	o / orderup        → ピックアップ (オーダーアップ)
//	oa / orderupalone  → ピックアップ (オーダーアップ + ゴーアローン)
//	pa / pass          → パス (フェーズに応じて PickUp or PassCall)
//	c / call <suit>    → スートをコール (1-4)
//	ca / callalone <suit> → スートをコール + ゴーアローン (1-4)
//	d / discard <i>    → カードを捨てる
//	p / play <i>       → カードをプレイ
//	n / next           → 次のトリックへ
//	nr / nextround     → 次のラウンドへ
//	sd / setdifficulty <0-2> → CPU難易度設定
//	sl / setlimit <n>  → ポイント上限設定
//	h / hint           → ヒント表示
//	log / l            → 棋譜表示
func (c *EuchreCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ei.GetConfig()
			return c.ei.ResetWithConfig(cfg)
		},
		[]string{
			"o", "orderup", "oa", "orderupalone",
			"pa", "pass",
			"c", "call", "ca", "callalone",
			"d", "discard", "p", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "o", "orderup":
				return c.ei.PickUp(true, false), true
			case "oa", "orderupalone":
				return c.ei.PickUp(true, true), true
			case "pa", "pass":
				return c.ei.Pass(), true
			case "c", "call":
				return cuiutil.WithParsedInt(args, "Suit is required (1-4).", "Invalid suit: %s.", 1, 4, func(suit int) string {
					return c.ei.CallTrump(suit, false)
				})
			case "ca", "callalone":
				return cuiutil.WithParsedInt(args, "Suit is required (1-4).", "Invalid suit: %s.", 1, 4, func(suit int) string {
					return c.ei.CallTrump(suit, true)
				})
			case "d", "discard":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ei.Discard)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ei.Play)
			case "n", "next":
				return c.ei.NextTrick(), true
			case "nr", "nextround":
				return c.ei.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.ei.GetConfig()
					cfg.CpuDifficulty = domain.EuchreCpuDifficulty(v)
					return c.ei.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedInt(args, "Point limit is required.", "Invalid point limit: %s. Please enter 1 or more.", 1, math.MaxInt, func(v int) string {
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
