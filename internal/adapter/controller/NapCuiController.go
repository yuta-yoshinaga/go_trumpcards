//go:build !js || !wasm || classic

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NapCuiController ナップのCUIコントローラークラス
type NapCuiController struct {
	di usecase.NapInteractorIF
}

// NewNapCuiController コンストラクタ
func NewNapCuiController(di usecase.NapInteractorIF) *NapCuiController {
	return &NapCuiController{di: di}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	b / bid <0/2/3/4/5>      → 入札 (0=Pass, 2/3/4=トリック数, 5=Nap)
//	pass                     → パス (入札 0)
//	<n> / play <n>           → カードをプレイ (プレイフェーズ)
//	n / next                 → 次のトリックへ
//	nr / nextround           → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *NapCuiController) Exec(command string) string {
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
				return cuiutil.WithParsedInt(args, "Bid is required (0=Pass, 2/3/4=tricks, 5=Nap).",
					"Invalid bid: %s. Please enter 0, 2, 3, 4 or 5.",
					int(domain.NapBidPass), int(domain.NapBidNap), c.di.Bid)
			case "pass":
				return c.di.Bid(int(domain.NapBidPass)), true
			case "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.di.Play)
			case "n", "next":
				return c.di.NextTrick(), true
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuDifficulty = domain.NapCpuDifficulty(v)
					return c.di.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}
