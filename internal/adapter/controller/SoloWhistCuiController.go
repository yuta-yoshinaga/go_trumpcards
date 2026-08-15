//go:build !js || !wasm || classic

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SoloWhistCuiController ソロ・ホイストのCUIコントローラークラス
type SoloWhistCuiController struct {
	di usecase.SoloWhistInteractorIF
}

// NewSoloWhistCuiController コンストラクタ
func NewSoloWhistCuiController(di usecase.SoloWhistInteractorIF) *SoloWhistCuiController {
	return &SoloWhistCuiController{di: di}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	b / bid <0-3>            → 入札 (0=Pass,1=Solo,2=Misère,3=Abundance)
//	pass                     → パス (入札 0)
//	<n> / play <n>           → カードをプレイ (プレイフェーズ)
//	n / next                 → 次のトリックへ
//	nr / nextround           → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *SoloWhistCuiController) Exec(command string) string {
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
				return cuiutil.WithParsedIntKeys(args, "bidRequiredAbundance", "invalidBid03",
					int(domain.SoloWhistBidPass), int(domain.SoloWhistBidAbundance), c.di.Bid)
			case "pass":
				return c.di.Bid(int(domain.SoloWhistBidPass)), true
			case "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.di.Play)
			case "n", "next":
				return c.di.NextTrick(), true
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuDifficulty = domain.SoloWhistCpuDifficulty(v)
					return c.di.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}
