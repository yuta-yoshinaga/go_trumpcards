//go:build !js || !wasm || extra2

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// IsraeliWhistCuiController イスラエリホイストCUIコントローラークラス
type IsraeliWhistCuiController struct {
	wi usecase.IsraeliWhistInteractorIF
}

// NewIsraeliWhistCuiController コンストラクタ
func NewIsraeliWhistCuiController(wi usecase.IsraeliWhistInteractorIF) *IsraeliWhistCuiController {
	return &IsraeliWhistCuiController{wi: wi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit          → ゲーム終了 ("bye.")
//	r / reset         → ゲームリセット (設定保持)
//	a / auction <n> <suit> → オークションで入札する (n は 5 以上、suit は 1:♠ 2:♣ 3:♥ 4:♦)
//	pass              → オークションを降りる
//	b / bid <n>       → 目標トリック数を宣言する (0..13)
//	p / play <i>      → 手札の i 番目を出す
//	n / next          → 次のラウンドへ
//	g / giveup        → 投了
//	h / hint          → ヒント表示
//	log / l           → 棋譜表示
func (c *IsraeliWhistCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.wi.ResetWithConfig(c.wi.GetConfig()) },
		[]string{"a", "auction", "pass", "b", "bid", "p", "play", "n", "next", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "a", "auction":
				return c.auction(args)
			case "pass":
				return c.wi.AuctionPass(), true
			case "b", "bid":
				return cuiutil.WithParsedInt(args, "Bid is required.", "Invalid bid: %s.",
					0, domain.IsraeliWhistHandSize, c.wi.Bid)
			case "p", "play":
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax, c.wi.Play)
			case "n", "next":
				return c.wi.NextRound(), true
			case "g", "giveup":
				return c.wi.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.wi.Hint, c.wi.ActionLog)
			}
		},
	)
}

// auction 入札は数とスートの 2 引数を取る。**どちらも既定値で埋めない。**
func (c *IsraeliWhistCuiController) auction(args []string) (string, bool) {
	bid, errMsg, ok := cuiutil.ParseIntArg(args, "Bid is required.", "Invalid bid: %s.",
		domain.IsraeliWhistMinAuctionBid, domain.IsraeliWhistHandSize)
	if !ok {
		return errMsg, true
	}
	suit, errMsg, ok := cuiutil.ParseIntArg(args[1:], "Suit is required.", "Invalid suit: %s.",
		domain.CardDesignSpade, domain.CardDesignMax)
	if !ok {
		return errMsg, true
	}
	return c.wi.AuctionBid(bid, suit), true
}
