//go:build !js || !wasm || extra

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KoenigrufenCuiController ケーニッヒルーフェン (Königrufen) のCUIコントローラークラス
type KoenigrufenCuiController struct {
	di usecase.KoenigrufenInteractorIF
}

// NewKoenigrufenCuiController コンストラクタ
func NewKoenigrufenCuiController(di usecase.KoenigrufenInteractorIF) *KoenigrufenCuiController {
	return &KoenigrufenCuiController{di: di}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                          → ゲーム終了 ("bye.")
//	r / reset                         → ゲームリセット (設定保持)
//	bid rufer                         → 入札
//	pass                              → パス
//	callking <1-4>                    → 呼ぶキングのスートを指名
//	discard <i0> ... <i5>             → 場札交換で 6 枚を捨てる
//	<n> / play <n>                    → カードをプレイ (プレイフェーズ)
//	n / next                          → 次のトリックへ
//	nr / nextround                    → 次のディールへ (スコアリング)
//	sd / setdifficulty <0-2>          → CPU難易度設定
//	h / hint                          → ヒント表示
//	log / l                           → 棋譜表示
func (c *KoenigrufenCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.di.GetConfig()
			return c.di.ResetWithConfig(cfg)
		},
		[]string{
			"bid", "pass", "callking", "discard", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "bid":
				return c.execBid(args)
			case "pass":
				return c.di.Pass(), true
			case "ck", "callking":
				return cuiutil.WithParsedIntKeys(args, "kingSuitRequired", "invalidSuitRange", 1, domain.KoenigrufenSuitCnt, c.di.CallKing)
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
					cfg.CpuDifficulty = domain.KoenigrufenCpuDifficulty(v)
					return c.di.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}

// execBid bid サブコマンドを解釈する。
func (c *KoenigrufenCuiController) execBid(args []string) (string, bool) {
	if len(args) == 0 {
		return invalidArg("bidRequiredRufer"), true
	}
	bid := koenigrufenParseBid(args[0])
	if bid == domain.KoenigrufenBidPass {
		return invalidArg("invalidBidRufer", "val", args[0]), true
	}
	return c.di.Bid(bid), true
}

// execDiscard discard サブコマンドを解釈する (6 枚のインデックス)。
func (c *KoenigrufenCuiController) execDiscard(args []string) (string, bool) {
	if len(args) < domain.KoenigrufenTalonSize {
		return invalidArg("sixIndicesRequiredDiscard"), true
	}
	indices, skipped := cuiutil.ParseIntSlice(args)
	if len(skipped) > 0 {
		return invalidArg("invalidCardIndex", "val", skipped[0]), true
	}
	return c.di.Discard(indices), true
}
