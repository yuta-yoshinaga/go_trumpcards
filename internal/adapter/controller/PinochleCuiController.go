package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// pinochleNoArgCommands maps no-arg CUI commands to Pinochle interactor methods.
var pinochleNoArgCommands = cuiutil.NewCommandMap[usecase.PinochleInteractorIF]().
	Add(usecase.PinochleInteractorIF.Pass, "pa", "pass").
	Add(usecase.PinochleInteractorIF.ConfirmMelds, "m", "meld").
	Add(usecase.PinochleInteractorIF.NextTrick, "n", "next").
	Add(usecase.PinochleInteractorIF.NextRound, "nr", "nextround").
	Add(usecase.PinochleInteractorIF.Hint, "h", "hint").
	Add(usecase.PinochleInteractorIF.ActionLog, "log", "l")

// pinochleArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var pinochleArgfulCommands = []string{
	"b", "bid", "t", "trump", "p", "play",
	"sd", "setdifficulty", "sl", "setlimit",
}

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
		append(pinochleNoArgCommands.Names(), pinochleArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := pinochleNoArgCommands.Lookup(cmd); ok {
				return fn(c.pi), true
			}
			switch cmd {
			case "b", "bid":
				return cuiutil.WithParsedIntKeys(args, "bidAmountRequired", "invalidBidAmount", domain.PinochleMinBid, math.MaxInt, c.pi.Bid)
			case "t", "trump":
				return cuiutil.WithParsedIntKeys(args, "suitRequiredRange", "invalidSuit", 1, 4, c.pi.CallTrump)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.pi.Play)
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
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
			default:
				return "", false
			}
		},
	)
}
