//go:build !js || !wasm || extra

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FrenchTarotCuiController フレンチタロット (French Tarot) のCUIコントローラークラス
type FrenchTarotCuiController struct {
	di usecase.FrenchTarotInteractorIF
}

// NewFrenchTarotCuiController コンストラクタ
func NewFrenchTarotCuiController(di usecase.FrenchTarotInteractorIF) *FrenchTarotCuiController {
	return &FrenchTarotCuiController{di: di}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                          → ゲーム終了 ("bye.")
//	r / reset                         → ゲームリセット (設定保持)
//	bid <petite|garde|gardesans|gardecontre>  → 入札
//	pass                              → パス
//	discard <i0> ... <i5>             → シアン交換で 6 枚を捨てる
//	<n> / play <n>                    → カードをプレイ (プレイフェーズ)
//	n / next                          → 次のトリックへ
//	nr / nextround                    → 次のディールへ (スコアリング)
//	sd / setdifficulty <0-2>          → CPU難易度設定
//	h / hint                          → ヒント表示
//	log / l                           → 棋譜表示
func (c *FrenchTarotCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.di.GetConfig()
			return c.di.ResetWithConfig(cfg)
		},
		[]string{
			"bid", "pass", "discard", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "bid":
				return c.execBid(args)
			case "pass":
				return c.di.Pass(), true
			case "discard":
				return c.execDiscard(args)
			case "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.di.Play)
			case "n", "next":
				return c.di.NextTrick(), true
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuDifficulty = domain.FrenchTarotCpuDifficulty(v)
					return c.di.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}

// execBid bid サブコマンドを解釈する。
func (c *FrenchTarotCuiController) execBid(args []string) (string, bool) {
	if len(args) == 0 {
		return "Bid is required (petite, garde, gardesans, or gardecontre).", true
	}
	bid := frenchTarotParseBid(args[0])
	if bid == domain.FrenchTarotBidPass {
		return "Invalid bid: " + args[0] + ". Please enter petite, garde, gardesans, or gardecontre.", true
	}
	return c.di.Bid(bid), true
}

// execDiscard discard サブコマンドを解釈する (6 枚のインデックス)。
func (c *FrenchTarotCuiController) execDiscard(args []string) (string, bool) {
	if len(args) < domain.FrenchTarotChienSize {
		return "Six card indices are required (e.g. discard 0 1 2 3 4 5).", true
	}
	indices, skipped := cuiutil.ParseIntSlice(args)
	if len(skipped) > 0 {
		return invalidArg("invalidCardIndex", "val", skipped[0]), true
	}
	return c.di.Discard(indices), true
}
