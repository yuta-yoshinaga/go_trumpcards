//go:build !js || !wasm || solo

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BidWhistCuiController Bid Whist CUIコントローラークラス
type BidWhistCuiController struct {
	bi usecase.BidWhistInteractorIF
}

// NewBidWhistCuiController コンストラクタ
func NewBidWhistCuiController(bi usecase.BidWhistInteractorIF) *BidWhistCuiController {
	return &BidWhistCuiController{bi: bi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit            → ゲーム終了 ("bye.")
//	r / reset           → ゲームリセット (設定保持)
//	b / bid <tricks> <dir> → ビッド (tricks 1-7, dir 0=Uptown 1=Downtown 2=NoTrump)
//	pa / pass           → パス
//	t / trump <suit>    → 切り札スート宣言 (1=S 2=C 3=H 4=D)
//	e / exchange <6 indices> → キティ交換 (捨てるカード6枚)
//	p / play <i>        → カードをプレイ
//	n / next            → 次のトリックへ
//	nr / nextround      → 次のラウンドへ
//	sd / setdifficulty <0-2> → CPU難易度設定
//	st / settarget <n>  → 勝利スコア設定
//	h / hint            → ヒント表示
//	log / l             → 棋譜表示
func (c *BidWhistCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.bi.GetConfig()
			return c.bi.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bid", "pa", "pass",
			"t", "trump", "e", "exchange", "p", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "st", "settarget",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				if len(args) < 2 {
					return "Usage: bid <tricks> <dir>  (tricks 1-7, dir 0=Uptown 1=Downtown 2=NoTrump)\n", true
				}
				tricks, errMsg, ok := cuiutil.ParseIntArgKeys(args[:1], "", "invalidTricks", domain.BidWhistMinBid, domain.BidWhistMaxBid)
				if !ok {
					return errMsg, true
				}
				dir, errMsg, ok := cuiutil.ParseIntArgKeys(args[1:2], "", "invalidDirection",
					domain.BidWhistDirectionUptown, domain.BidWhistDirectionNoTrump)
				if !ok {
					return errMsg, true
				}
				return c.bi.Bid(tricks, dir), true
			case "pa", "pass":
				return c.bi.Pass(), true
			case "t", "trump":
				return cuiutil.WithParsedIntKeys(args, "trumpSuitRequiredLettersPlain", "invalidSuitRange", 1, 4, func(s int) string {
					return c.bi.DeclareTrump(s)
				})
			case "e", "exchange":
				if len(args) < domain.BidWhistKittySize {
					return "Usage: exchange <i1> <i2> <i3> <i4> <i5> <i6>  (six card indices to discard)\n", true
				}
				idxs := make([]int, domain.BidWhistKittySize)
				for i := 0; i < domain.BidWhistKittySize; i++ {
					v, errMsg, ok := cuiutil.ParseIntArgKeys(args[i:i+1], "", "invalidCardIndex", 0, math.MaxInt)
					if !ok {
						return errMsg, true
					}
					idxs[i] = v
				}
				return c.bi.ExchangeKitty(idxs), true
			case "p", "play":
				cardIdx, errMsg, ok := cuiutil.ParseIntArgKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax)
				if !ok {
					return errMsg, true
				}
				return c.bi.Play(cardIdx), true
			case "n", "next":
				return c.bi.NextTrick(), true
			case "nr", "nextround":
				return c.bi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.bi.GetConfig()
					cfg.CpuDifficulty = domain.BidWhistCpuDifficulty(v)
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
