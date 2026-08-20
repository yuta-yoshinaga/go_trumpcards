//go:build !js || !wasm || extra3

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BridgeCuiController ブリッジCUIコントローラークラス
type BridgeCuiController struct {
	bi usecase.BridgeInteractorIF
}

// NewBridgeCuiController コンストラクタ
func NewBridgeCuiController(bi usecase.BridgeInteractorIF) *BridgeCuiController {
	return &BridgeCuiController{bi: bi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit           → ゲーム終了 ("bye.")
//	r / reset          → ゲームリセット (設定保持)
//	b / bid <type> [level] [suit] → ビッド (type: 0=Pass, 1=Normal, 2=Double, 3=Redouble)
//	p / play <i>       → カードをプレイ
//	n / next           → 次のトリックへ
//	nr / nextround     → 次のラウンドへ
//	sd / setdifficulty <0-2> → CPU難易度設定
//	h / hint           → ヒント表示
//	log / l            → 棋譜表示
func (c *BridgeCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.bi.GetConfig()
			return c.bi.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bid",
			"p", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				return cuiutil.WithParsedIntKeys(args, "bidTypeRequired", "invalidBidType", 0, 3, func(bidType int) string {
					bidLevel, errMsg, ok := cuiutil.ParseOptionalIntKeys(args, 1, 0, "invalidBidLevel")
					if !ok {
						return errMsg
					}
					bidSuit, errMsg, ok := cuiutil.ParseOptionalIntKeys(args, 2, 0, "invalidSuit")
					if !ok {
						return errMsg
					}
					return c.bi.Bid(bidType, bidLevel, bidSuit)
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
					cfg.CpuDifficulty = domain.BridgeCpuDifficulty(v)
					return c.bi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.bi.Hint, c.bi.ActionLog)
			}
		},
	)
}
