//go:build !js || !wasm || extra3

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HasenpfefferCuiController ハーゼンプフェファーCUIコントローラークラス
type HasenpfefferCuiController struct {
	hi usecase.HasenpfefferInteractorIF
}

// NewHasenpfefferCuiController コンストラクタ
func NewHasenpfefferCuiController(hi usecase.HasenpfefferInteractorIF) *HasenpfefferCuiController {
	return &HasenpfefferCuiController{hi: hi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit          → ゲーム終了 ("bye.")
//	r / reset         → ゲームリセット (設定保持)
//	b / bid <n>       → n トリック宣言する (3..6)
//	pass              → 降りる (親は3人が降りたら降りられない)
//	d / discard <i> <s> → 伏せ札を取り込んだあと i を捨てて切り札 s を宣言
//	p / play <i>      → 手札の i 番目を出す
//	n / next          → 次のハンドへ
//	g / giveup        → 投了
//	h / hint          → ヒント表示
//	log / l           → 棋譜表示
func (c *HasenpfefferCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.hi.ResetWithConfig(c.hi.GetConfig()) },
		[]string{"b", "bid", "pass", "d", "discard", "p", "play", "n", "next", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				return cuiutil.WithParsedInt(args, "Bid is required.", "Invalid bid: %s.",
					domain.HasenpfefferMinBid, domain.HasenpfefferMaxBid, c.hi.Bid)
			case "pass":
				// **降りるのは専用コマンド。** `bid 0` を通すと下限の検査が要らなく
				// なり、範囲外の宣言まですり抜ける。
				return c.hi.Bid(0), true
			case "d", "discard":
				return c.discard(args)
			case "p", "play":
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax, c.hi.Play)
			case "n", "next":
				return c.hi.NextHand(), true
			case "g", "giveup":
				return c.hi.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.hi.Hint, c.hi.ActionLog)
			}
		},
	)
}

// discard は捨てる札のインデックスと切り札スートの 2 引数を取る。
//
// **どちらも既定値で埋めない。** 埋めると捨てていない札が捨てられたり、
// 選んでいないスートが切り札になる。
func (c *HasenpfefferCuiController) discard(args []string) (string, bool) {
	idx, errMsg, ok := cuiutil.ParseIntArg(args, "Card index is required.", "Invalid card index: %s.",
		cuiutil.NoMin, cuiutil.NoMax)
	if !ok {
		return errMsg, true
	}
	suit, errMsg, ok := cuiutil.ParseIntArg(args[1:], "Suit is required.", "Invalid suit: %s.",
		domain.CardDesignSpade, domain.CardDesignMax)
	if !ok {
		return errMsg, true
	}
	return c.hi.Discard(idx, suit), true
}
