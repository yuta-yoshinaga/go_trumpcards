package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PinochleCuiController ピノクルCUIコントローラークラス
type PinochleCuiController struct {
	pi usecase.PinochleInteractorIF
}

// NewPinochleCuiController コンストラクタ
func NewPinochleCuiController(pi usecase.PinochleInteractorIF) *PinochleCuiController {
	return &PinochleCuiController{pi: pi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit           → ゲーム終了 ("bye.")
//	r / reset          → ゲームリセット (設定保持)
//	b / bid <amount>   → ビッド
//	pa / pass          → パス
//	t / trump <suit>   → トランプスート宣言 (1-4)
//	m / meld           → メルド確認
//	p / play <i>       → カードをプレイ
//	n / next           → 次のトリックへ
//	nr / nextround     → 次のラウンドへ
//	sd / setdifficulty <0-2> → CPU難易度設定
//	sl / setlimit <n>  → ポイント上限設定
//	h / hint           → ヒント表示
//	log / l            → 棋譜表示
func (c *PinochleCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.pi.GetConfig()
			return c.pi.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bid", "pa", "pass",
			"t", "trump", "m", "meld",
			"p", "play", "n", "next",
			"nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				return cuiutil.WithParsedInt(args, "Bid amount is required.", "Invalid bid amount: %s.", domain.PinochleMinBid, math.MaxInt, c.pi.Bid)
			case "pa", "pass":
				return c.pi.Pass(), true
			case "t", "trump":
				return cuiutil.WithParsedInt(args, "Suit is required (1-4).", "Invalid suit: %s.", 1, 4, c.pi.CallTrump)
			case "m", "meld":
				return c.pi.ConfirmMelds(), true
			case "p", "play":
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax, c.pi.Play)
			case "n", "next":
				return c.pi.NextTrick(), true
			case "nr", "nextround":
				return c.pi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.pi.GetConfig()
					cfg.CpuDifficulty = domain.PinochleCpuDifficulty(v)
					return c.pi.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedInt(args, "Point limit is required.", "Invalid point limit: %s.", 1, math.MaxInt, func(v int) string {
					cfg := c.pi.GetConfig()
					cfg.PointLimit = v
					return c.pi.ResetWithConfig(cfg)
				})
			case "h", "hint":
				return c.pi.Hint(), true
			case "log", "l":
				return c.pi.ActionLog(), true
			default:
				return "", false
			}
		},
	)
}
