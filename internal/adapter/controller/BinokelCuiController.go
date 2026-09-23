package controller

import (
	"math"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// binokelNoArgCommands maps no-arg CUI commands to Binokel interactor methods.
var binokelNoArgCommands = cuiutil.NewCommandMap[usecase.BinokelInteractorIF]().
	Add(usecase.BinokelInteractorIF.Pass, "pa", "pass").
	Add(usecase.BinokelInteractorIF.ConfirmMelds, "m", "meld").
	Add(usecase.BinokelInteractorIF.NextTrick, "n", "next").
	Add(usecase.BinokelInteractorIF.NextRound, "nr", "nextround").
	Add(usecase.BinokelInteractorIF.Hint, "h", "hint").
	Add(usecase.BinokelInteractorIF.ActionLog, "log", "l")

// binokelArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var binokelArgfulCommands = []string{
	"b", "bid", "d", "discard", "t", "trump", "p", "play",
	"sd", "setdifficulty", "sl", "setlimit",
}

// BinokelCuiController ビノクルCUIコントローラークラス
type BinokelCuiController struct {
	pi usecase.BinokelInteractorIF
}

// NewBinokelCuiController コンストラクタ
func NewBinokelCuiController(pi usecase.BinokelInteractorIF) *BinokelCuiController {
	return &BinokelCuiController{pi: pi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	b / bid <amount>         → ビッド (150以上・10刻み)
//	pa / pass                → パス
//	d / discard <i> <j> <k>  → Dabbへカード3枚を捨てる
//	t / trump <suit>         → トランプスート宣言 (1-4)
//	m / meld                 → メルド確認
//	p / play <i>             → カードをプレイ
//	n / next                 → 次のトリックへ
//	nr / nextround           → 次のラウンドへ
//	sd / setdifficulty <0-2> → CPU難易度設定
//	sl / setlimit <n>        → ポイント上限設定
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *BinokelCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.pi.GetConfig()
			return c.pi.ResetWithConfig(cfg)
		},
		append(binokelNoArgCommands.Names(), binokelArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := binokelNoArgCommands.Lookup(cmd); ok {
				return fn(c.pi), true
			}
			switch cmd {
			case "b", "bid":
				return cuiutil.WithParsedIntKeys(args, "bidAmountRequired", "invalidBidAmount", domain.BinokelMinBid, math.MaxInt, func(v int) string {
					if (v-domain.BinokelMinBid)%domain.BinokelBidStep != 0 {
						return invalidArg("invalidBidAmount", "val", strconv.Itoa(v))
					}
					return c.pi.Bid(v)
				})
			case "d", "discard":
				if len(args) < 3 {
					return invalidArg("usageDiscardIJKThreeCardIndices"), true
				}
				idxA, errMsg, ok := cuiutil.ParseIntArgKeys(args[:1], "", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax)
				if !ok {
					return errMsg, true
				}
				idxB, errMsg, ok := cuiutil.ParseIntArgKeys(args[1:2], "", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax)
				if !ok {
					return errMsg, true
				}
				idxC, errMsg, ok := cuiutil.ParseIntArgKeys(args[2:3], "", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax)
				if !ok {
					return errMsg, true
				}
				return c.pi.DiscardToDabb([]int{idxA, idxB, idxC}), true
			case "t", "trump":
				return cuiutil.WithParsedIntKeys(args, "suitRequiredRange", "invalidSuit", 1, 4, c.pi.CallTrump)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.pi.Play)
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.pi.GetConfig()
					cfg.CpuDifficulty = domain.BinokelCpuDifficulty(v)
					return c.pi.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedIntKeys(args, "pointLimitRequired", "invalidPointLimitPlain", 1, math.MaxInt, func(v int) string {
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
