package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OhHellCuiController オー・ヘルCUIコントローラークラス
type OhHellCuiController struct {
	oi usecase.OhHellInteractorIF
}

// NewOhHellCuiController コンストラクタ
func NewOhHellCuiController(oi usecase.OhHellInteractorIF) *OhHellCuiController {
	return &OhHellCuiController{oi: oi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit         → ゲーム終了 ("bye.")
//	r / reset        → ゲームリセット (設定保持)
//	b / bid <n>      → ビッドを宣言
//	p / play <i>     → カードをプレイ
//	n / next         → 次のトリックへ
//	nr / nextround   → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	sm / setmaxhand <n> → 最大手札枚数設定
//	h / hint         → ヒント表示
//	log / l          → 棋譜表示
func (c *OhHellCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.oi.GetConfig()
			return c.oi.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bid", "p", "play", "n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sm", "setmaxhand", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				return cuiutil.WithParsedInt(args, "Bid value is required.", "Invalid bid value: %s.", 0, cuiutil.NoMax, c.oi.Bid)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.oi.Play)
			case "n", "next":
				return c.oi.NextTrick(), true
			case "nr", "nextround":
				return c.oi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.oi.GetConfig()
					cfg.CpuDifficulty = domain.OhHellCpuDifficulty(v)
					return c.oi.ResetWithConfig(cfg)
				})
			case "sm", "setmaxhand":
				return cuiutil.WithParsedInt(args, "Max hand size is required (1-13).", "Invalid max hand size: %s. Please enter 1-13.", 1, 13, func(v int) string {
					cfg := c.oi.GetConfig()
					cfg.MaxHandSize = v
					return c.oi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.oi.Hint, c.oi.ActionLog)
			}
		},
	)
}
