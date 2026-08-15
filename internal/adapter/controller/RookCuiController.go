//go:build !js || !wasm || extra3

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RookCuiController ルーク(Rook) CUIコントローラークラス
type RookCuiController struct {
	fi usecase.RookInteractorIF
}

// NewRookCuiController コンストラクタ
func NewRookCuiController(fi usecase.RookInteractorIF) *RookCuiController {
	return &RookCuiController{fi: fi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit               → ゲーム終了 ("bye.")
//	r / reset              → ゲームリセット (設定保持)
//	b / bid <points>       → ビッド (70-120, 5刻み)
//	pa / pass              → パス
//	e / exchange <i>...<m> <color> → ネスト交換 (捨てる5枚 + 切り札色 1-4)
//	p / play <i>           → カードをプレイ
//	n / next               → 次のトリックへ
//	nr / nextround         → 次のラウンドへ
//	sd / setdifficulty <0-2> → CPU難易度設定
//	st / settarget <n>     → 勝利スコア設定
//	h / hint               → ヒント表示
//	log / l                → 棋譜表示
func (c *RookCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.fi.GetConfig()
			return c.fi.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bid", "pa", "pass",
			"e", "exchange", "p", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "st", "settarget",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				return cuiutil.WithParsedIntKeys(args, "bidRequired70120", "invalidBid",
					domain.RookMinBid, domain.RookMaxBid, func(v int) string {
						return c.fi.Bid(v)
					})
			case "pa", "pass":
				return c.fi.Pass(), true
			case "e", "exchange":
				if len(args) < 6 {
					return invalidArg("usageExchangeIJKLMColor5CardIndicesTrumpColor14"), true
				}
				idxs := make([]int, 5)
				for i := 0; i < 5; i++ {
					v, errMsg, ok := cuiutil.ParseIntArgKeys(args[i:i+1], "", "invalidCardIndex", 0, math.MaxInt)
					if !ok {
						return errMsg, true
					}
					idxs[i] = v
				}
				color, errMsg, ok := cuiutil.ParseIntArgKeys(args[5:6], "", "invalidTrumpColor", 1, 4)
				if !ok {
					return errMsg, true
				}
				return c.fi.ExchangeNest(idxs, color), true
			case "p", "play":
				cardIdx, errMsg, ok := cuiutil.ParseIntArgKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax)
				if !ok {
					return errMsg, true
				}
				return c.fi.Play(cardIdx), true
			case "n", "next":
				return c.fi.NextTrick(), true
			case "nr", "nextround":
				return c.fi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.fi.GetConfig()
					cfg.CpuDifficulty = domain.RookCpuDifficulty(v)
					return c.fi.ResetWithConfig(cfg)
				})
			case "st", "settarget":
				return cuiutil.WithParsedIntKeys(args, "targetScoreRequired", "invalidTargetScore", 1, math.MaxInt, func(v int) string {
					cfg := c.fi.GetConfig()
					cfg.TargetScore = v
					return c.fi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.fi.Hint, c.fi.ActionLog)
			}
		},
	)
}
