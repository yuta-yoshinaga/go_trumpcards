//go:build !js || !wasm || extra

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CinchCuiController はチンチ (Cinch) の CUI コントローラークラス。
type CinchCuiController struct {
	ci usecase.CinchInteractorIF
}

// NewCinchCuiController コンストラクタ。
func NewCinchCuiController(ci usecase.CinchInteractorIF) *CinchCuiController {
	return &CinchCuiController{ci: ci}
}

// Exec コマンド実行。
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	bid <0-14>               → ビッド (0=pass)
//	t / trump <1-4>          → 切り札スートを宣言 (1=Spade 2=Clover 3=Heart 4=Diamond)
//	p / play <n>             → カードをプレイ
//	nr / nextround / n       → 次のディールへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *CinchCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"bid", "t", "trump", "p", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "bid":
				return cuiutil.WithParsedIntKeys(args, "bidValueRequiredPass114", "invalidBid014", domain.CinchPassBid, domain.CinchMaxBid, c.ci.Bid)
			case "t", "trump":
				return cuiutil.WithParsedIntKeys(args, "trumpSuitRequiredRange", "invalidTrumpSuitRange", domain.CardDesignSpade, domain.CardDesignDiamond, c.ci.NameTrump)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", 0, domain.CinchHandSize-1, c.ci.Play)
			case "n", "next", "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.CinchCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.ci.Hint, c.ci.ActionLog)
			}
		},
	)
}
