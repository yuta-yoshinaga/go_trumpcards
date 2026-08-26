//go:build !js || !wasm || casino

package controller

import (
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HorseCuiController は H.O.R.S.E. の CUI コントローラー。
type HorseCuiController struct {
	hi usecase.HorseInteractorIF
}

// NewHorseCuiController コンストラクタ。
func NewHorseCuiController(hi usecase.HorseInteractorIF) *HorseCuiController {
	return &HorseCuiController{hi: hi}
}

// Exec コマンド実行。
//
//	q / quit                → ゲーム終了 ("bye.")
//	r / reset               → ゲームリセット (設定保持)
//	f / fold                → フォールド
//	x / check               → チェック
//	c / call                → コール
//	b / bet <n>             → ベット
//	raise <n>               → レイズ
//	allin                   → オールイン
//	d / draw <n...>         → 引き直し (2-7 トリプルドローのみ、0 始まり)
//	sp / stand              → スタンドパット (引かない)
//	n / next                → 次のハンドへ
//	ss / setseats <4|6|9>   → 席数設定
//	sh / sethands <1-10>    → 1 種目あたりのハンド数設定
//	h / hint                → ヒント表示
//	log / l                 → 棋譜表示
func (c *HorseCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.hi.ResetWithConfig(c.hi.GetConfig())
		},
		[]string{
			"f", "fold", "x", "check", "c", "call", "b", "bet", "raise", "allin",
			"d", "draw", "sp", "stand",
			"n", "next", "ss", "setseats", "sh", "sethands", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "f", "fold":
				return c.hi.Action(domain.HoldemActionFold, 0, 0), true
			case "x", "check":
				return c.hi.Action(domain.HoldemActionCheck, 0, 0), true
			case "c", "call":
				return c.hi.Action(domain.HoldemActionCall, 0, 0), true
			case "b", "bet":
				return cuiutil.WithParsedIntKeys(args, "betAmountRequired", "invalidAmount", 1, cuiutil.NoMax, func(v int) string {
					return c.hi.Action(domain.HoldemActionBet, v, 0)
				})
			case "raise":
				return cuiutil.WithParsedIntKeys(args, "raiseAmountRequired", "invalidAmount", 1, cuiutil.NoMax, func(v int) string {
					return c.hi.Action(domain.HoldemActionRaise, v, 0)
				})
			case "allin":
				return c.hi.Action(domain.HoldemActionAllIn, 0, 0), true
			case "d", "draw":
				return c.execDraw(args)
			case "sp", "stand":
				return c.hi.Exchange(nil), true
			case "n", "next":
				return c.hi.NextHand(), true
			case "ss", "setseats":
				return c.execSetSeats(args)
			case "sh", "sethands":
				return cuiutil.WithParsedIntKeys(args, "numberOfHandsPerDisciplineRequired110", "invalidNumberOfHands110",
					domain.HorseMinHandsPerDiscipline, domain.HorseMaxHandsPerDiscipline,
					func(v int) string {
						cfg := c.hi.GetConfig()
						cfg.HandsPerDiscipline = v
						return c.hi.ResetWithConfig(cfg)
					})
			default:
				return handleCuiHintAndLog(cmd, c.hi.Hint, c.hi.ActionLog)
			}
		},
	)
}

// execDraw は引き直しを渡す。
//
// **番号は 0 始まり。** 引き直しのある他のゲーム (2-7 単体・ムス) が `d 0 2`
// と数えるので、ここだけ 1 始まりにすると同じ卓の同じ操作が 1 枚ずれる。
// 画面もドローの番だけ `[0]♠5` の形で番号を振る。
//
// **打ち間違いは打つ前に断る。** 1 つでも読めない番号があれば 1 枚も捨てない ──
// 残りを「別の合法な手」として打ってしまうと、取り返しがつかない (#5390)。
func (c *HorseCuiController) execDraw(args []string) (string, bool) {
	if len(args) == 0 {
		// 引数無しは「引かない」ではなく打ち間違い。スタンドパットは sp。
		return invalidArg("invalidCardIndexUsageD"), true
	}
	indices, skipped := cuiutil.ParseBoundedIntSlice(args, 0, domain.DeuceToSevenHandSize-1)
	if len(skipped) > 0 {
		return invalidArg("invalidCardIndexPlain", "val", strings.Join(skipped, ", ")), true
	}
	return c.hi.Exchange(indices), true
}

// execSetSeats は席数を変える。
//
// **選べる数はバリアントで違う。** H.O.R.S.E. は 4 / 6 / 9 だが、
// Eight-Game Mix は 4 人卓しか作れない (2-7 トリプルドローが 4 席まで) ──
// 6 を通すと、6 種目目で理由も出さずにマッチが終わる。
func (c *HorseCuiController) execSetSeats(args []string) (string, bool) {
	if len(args) == 0 {
		return invalidArg("numberOfSeatsRequired469"), true
	}
	cfg := c.hi.GetConfig()
	return cuiutil.WithParsedIntKeys(args, "numberOfSeatsRequired469", "invalidNumberOfSeats46Or9",
		cuiutil.NoMin, cuiutil.NoMax, func(v int) string {
			cfg.Seats = v
			if err := cfg.Validate(); err != nil {
				return invalidArg("invalidNumberOfSeats469")
			}
			return c.hi.ResetWithConfig(cfg)
		})
}
