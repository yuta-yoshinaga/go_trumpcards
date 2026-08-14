//go:build !js || !wasm || casino

package controller

import (
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
				return cuiutil.WithParsedInt(args, "Bet amount is required.",
					"Invalid amount: %s.", 1, cuiutil.NoMax, func(v int) string {
						return c.hi.Action(domain.HoldemActionBet, v, 0)
					})
			case "raise":
				return cuiutil.WithParsedInt(args, "Raise amount is required.",
					"Invalid amount: %s.", 1, cuiutil.NoMax, func(v int) string {
						return c.hi.Action(domain.HoldemActionRaise, v, 0)
					})
			case "allin":
				return c.hi.Action(domain.HoldemActionAllIn, 0, 0), true
			case "n", "next":
				return c.hi.NextHand(), true
			case "ss", "setseats":
				return c.execSetSeats(args)
			case "sh", "sethands":
				return cuiutil.WithParsedInt(args,
					"Number of hands per discipline is required (1-10).",
					"Invalid number of hands: %s. Please enter 1-10.",
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

// execSetSeats は席数を変える。
//
// **選べるのは 4 / 6 / 9 だけ。** 種目側の卓サイズと同じものしか作れないので、
// 範囲で受けると 5 のような作れない数が通ってしまう。
func (c *HorseCuiController) execSetSeats(args []string) (string, bool) {
	if len(args) == 0 {
		return "Number of seats is required (4, 6 or 9).", true
	}
	return cuiutil.WithParsedInt(args, "Number of seats is required (4, 6 or 9).",
		"Invalid number of seats: %s. Please enter 4, 6 or 9.",
		cuiutil.NoMin, cuiutil.NoMax, func(v int) string {
			if !domain.HorseValidSeats(v) {
				return "Invalid number of seats: please enter 4, 6 or 9."
			}
			cfg := c.hi.GetConfig()
			cfg.Seats = v
			return c.hi.ResetWithConfig(cfg)
		})
}
