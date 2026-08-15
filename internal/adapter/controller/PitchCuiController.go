package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PitchCuiController ピッチCUIコントローラークラス
type PitchCuiController struct {
	pi usecase.PitchInteractorIF
}

// NewPitchCuiController コンストラクタ
func NewPitchCuiController(pi usecase.PitchInteractorIF) *PitchCuiController {
	return &PitchCuiController{pi: pi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit               → ゲーム終了 ("bye.")
//	r / reset              → ゲームリセット (設定保持)
//	b / bid <n>            → ビッドを宣言 (0=pass, 2-4)
//	p / play <i>           → カードをプレイ
//	n / next               → 次のトリックへ
//	nr / nextround         → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	sl / setlimit <n>      → ポイント上限設定
//	h / hint               → ヒント表示
//	log / l                → 棋譜表示
func (c *PitchCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.pi.GetConfig()
			return c.pi.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bid", "p", "play", "n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				return cuiutil.WithParsedInt(args, "Bid value is required (0=pass, 2-4).", "Invalid bid value: %s.", domain.PitchPassBid, domain.PitchMaxBid, c.pi.Bid)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.pi.Play)
			case "n", "next":
				return c.pi.NextTrick(), true
			case "nr", "nextround":
				return c.pi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.pi.GetConfig()
					cfg.CpuDifficulty = domain.PitchCpuDifficulty(v)
					return c.pi.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedInt(args, "Point limit is required.", "Invalid point limit: %s. Please enter 1 or more.", 1, math.MaxInt, func(v int) string {
					cfg := c.pi.GetConfig()
					cfg.PointLimit = v
					return c.pi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.pi.Hint, c.pi.ActionLog)
			}
		},
	)
}
