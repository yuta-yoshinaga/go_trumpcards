//go:build !js || !wasm || casino

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TwentyNineCuiController トゥエンティナイン (29) のCUIコントローラークラス
type TwentyNineCuiController struct {
	di usecase.TwentyNineInteractorIF
}

// NewTwentyNineCuiController コンストラクタ
func NewTwentyNineCuiController(di usecase.TwentyNineInteractorIF) *TwentyNineCuiController {
	return &TwentyNineCuiController{di: di}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                    → ゲーム終了 ("bye.")
//	r / reset                   → ゲームリセット (設定保持)
//	b / bid <0/16/20/24/28>     → 入札 (0=Pass)
//	pass                        → パス (入札 0)
//	<n> / play <n>              → カードをプレイ (プレイフェーズ)
//	n / next                    → 次のトリックへ
//	nr / nextround              → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2>    → CPU難易度設定
//	h / hint                    → ヒント表示
//	log / l                     → 棋譜表示
func (c *TwentyNineCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.di.GetConfig()
			return c.di.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bid", "pass", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				return cuiutil.WithParsedIntKeys(args, "bidRequired16", "invalidBid16",
					int(domain.TwentyNineBidPass), int(domain.TwentyNineBidTwentyEight), c.di.Bid)
			case "pass":
				return c.di.Bid(int(domain.TwentyNineBidPass)), true
			case "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.di.Play)
			case "n", "next":
				return c.di.NextTrick(), true
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuDifficulty = domain.TwentyNineCpuDifficulty(v)
					return c.di.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}
