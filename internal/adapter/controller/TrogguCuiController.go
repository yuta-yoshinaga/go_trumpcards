//go:build !js || !wasm || extra

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TrogguCuiController トロッグの CUI コントローラー。
type TrogguCuiController struct {
	ti usecase.TrogguInteractorIF
}

// NewTrogguCuiController コンストラクタ。
func NewTrogguCuiController(ti usecase.TrogguInteractorIF) *TrogguCuiController {
	return &TrogguCuiController{ti: ti}
}

// Exec コマンド実行。
//
//	q / quit                        → ゲーム終了 ("bye.")
//	r / reset                       → ゲームリセット (設定保持)
//	bid trois|solo|piccolo|misere   → 入札
//	pass                            → パス
//	play <n>                        → カードをプレイ
//	n / next                        → 次のトリックへ
//	nr / nextround                  → 次のディールへ
//	sd / setdifficulty <0-2>        → CPU 難易度設定
//	st / setdeals <1-12>            → ディール数設定
//	h / hint                        → ヒント表示
//	log / l                         → 棋譜表示
func (c *TrogguCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.ti.ResetWithConfig(c.ti.GetConfig())
		},
		[]string{
			"bid", "pass", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "st", "setdeals", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "bid":
				return c.execBid(args)
			case "pass":
				return c.ti.Pass(), true
			case "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired",
					"invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ti.Play)
			case "n", "next":
				return c.ti.NextTrick(), true
			case "nr", "nextround":
				return c.ti.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args,
					"cpuDifficultyRequired",
					"invalidCpuDifficulty", 0, 2, func(v int) string {
						cfg := c.ti.GetConfig()
						cfg.CpuDifficulty = domain.TrogguCpuDifficulty(v)
						return c.ti.ResetWithConfig(cfg)
					})
			case "st", "setdeals":
				return cuiutil.WithParsedIntKeys(args, "numberOfDealsRequired112", "invalidNumberOfDeals112",
					domain.TrogguMinDeals, domain.TrogguMaxDeals, func(v int) string {
						cfg := c.ti.GetConfig()
						cfg.TargetDeals = v
						return c.ti.ResetWithConfig(cfg)
					})
			default:
				return handleCuiHintAndLog(cmd, c.ti.Hint, c.ti.ActionLog)
			}
		},
	)
}

// execBid bid サブコマンドを解釈する。
func (c *TrogguCuiController) execBid(args []string) (string, bool) {
	if len(args) == 0 {
		return "Bid is required (trois, solo, piccolo or misere).", true
	}
	bid := trogguParseBid(args[0])
	if bid == domain.TrogguBidPass {
		return "Invalid bid: " + args[0] + ". Please enter trois, solo, piccolo or misere.", true
	}
	return c.ti.Bid(bid), true
}
