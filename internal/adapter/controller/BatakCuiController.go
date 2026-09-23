package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BatakCuiController Batak CUI コントローラークラス
type BatakCuiController struct {
	ci usecase.BatakInteractorIF
}

// NewBatakCuiController コンストラクタ
func NewBatakCuiController(ci usecase.BatakInteractorIF) *BatakCuiController {
	return &BatakCuiController{ci: ci}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit               → ゲーム終了 ("bye.")
//	r / reset              → ゲームリセット (設定保持)
//	b / bid <n>            → ビッドを宣言 (5〜13、0=パス)
//	pass                   → パス (0 を宣言)
//	p / play <i>           → カードをプレイ
//	n / next               → 次のトリックへ
//	nr / nextround         → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU 難易度設定
//	sr / setrounds <n>     → 総ラウンド数設定 (1 以上)
//	h / hint               → ヒント表示
//	log / l                → 棋譜表示
func (c *BatakCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bid", "pass", "p", "play", "n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sr", "setrounds", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				return cuiutil.WithParsedIntKeys(args, "bidValueRequiredPass513", "invalidBidValue", domain.BatakPassBid, domain.BatakMaxBid, c.ci.Bid)
			case "pass":
				return c.ci.Bid(domain.BatakPassBid), true
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Play)
			case "n", "next":
				return c.ci.NextTrick(), true
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.BatakCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sr", "setrounds":
				return cuiutil.WithParsedIntKeys(args, "totalRoundsRequired", "invalidRoundCount1OrMore", 1, math.MaxInt, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.MaxRounds = v
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.ci.Hint, c.ci.ActionLog)
			}
		},
	)
}
