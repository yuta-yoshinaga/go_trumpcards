package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CallBreakCuiController Call Break CUI コントローラークラス
type CallBreakCuiController struct {
	ci usecase.CallBreakInteractorIF
}

// NewCallBreakCuiController コンストラクタ
func NewCallBreakCuiController(ci usecase.CallBreakInteractorIF) *CallBreakCuiController {
	return &CallBreakCuiController{ci: ci}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit               → ゲーム終了 ("bye.")
//	r / reset              → ゲームリセット (設定保持)
//	b / bid <n>            → ビッドを宣言 (1〜13)
//	p / play <i>           → カードをプレイ
//	n / next               → 次のトリックへ
//	nr / nextround         → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU 難易度設定
//	sr / setrounds <n>     → 総ラウンド数設定 (1 以上)
//	h / hint               → ヒント表示
//	log / l                → 棋譜表示
func (c *CallBreakCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bid", "p", "play", "n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sr", "setrounds", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				return cuiutil.WithParsedInt(args, "Bid value is required (1-13).", "Invalid bid value: %s.", domain.CallBreakMinBid, domain.CallBreakHandSize, c.ci.Bid)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Play)
			case "n", "next":
				return c.ci.NextTrick(), true
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.CallBreakCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sr", "setrounds":
				return cuiutil.WithParsedInt(args, "Total rounds is required.", "Invalid round count: %s. Please enter 1 or more.", 1, math.MaxInt, func(v int) string {
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
