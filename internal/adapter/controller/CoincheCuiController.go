//go:build !js || !wasm || extra3

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CoincheCuiController コワンシュCUIコントローラークラス
type CoincheCuiController struct {
	bi usecase.CoincheInteractorIF
}

// NewCoincheCuiController コンストラクタ
func NewCoincheCuiController(bi usecase.CoincheInteractorIF) *CoincheCuiController {
	return &CoincheCuiController{bi: bi}
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
func (c *CoincheCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.bi.GetConfig()
			return c.bi.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bid",
			"pa", "pass",
			"co", "coinche",
			"su", "surcoinche",
			"ok",
			"p", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "st", "settarget",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				// **点とスートは 2 つで 1 つの宣言。** どちらかに既定値を
				// 置くと、打ち間違いが別の契約として黙って通る。
				points, msg, ok := cuiutil.ParseIntArgKeys(args, "coincheBidUsage", "coincheBidUsage",
					domain.CoincheContractPoints[0], domain.CoincheCapotPoints)
				if !ok {
					return msg, true
				}
				suit, msg, ok := cuiutil.ParseIntArgKeys(args[1:], "coincheBidUsage", "invalidSuit",
					domain.CardDesignSpade, domain.CardDesignMax)
				if !ok {
					return msg, true
				}
				return c.bi.Bid(points, suit), true
			case "pa", "pass":
				return c.bi.Pass(), true
			case "co", "coinche":
				return c.bi.Coinche(), true
			case "su", "surcoinche":
				return c.bi.Surcoinche(), true
			case "ok":
				return c.bi.DeclineDouble(), true
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.bi.Play)
			case "n", "next":
				return c.bi.NextTrick(), true
			case "nr", "nextround":
				return c.bi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.bi.GetConfig()
					cfg.CpuDifficulty = domain.CoincheCpuDifficulty(v)
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
