//go:build !js || !wasm || extra2

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BidEuchreCuiController ビッド・ユーカー (Bid Euchre) CUIコントローラークラス
type BidEuchreCuiController struct {
	bi usecase.BidEuchreInteractorIF
}

// NewBidEuchreCuiController コンストラクタ
func NewBidEuchreCuiController(bi usecase.BidEuchreInteractorIF) *BidEuchreCuiController {
	return &BidEuchreCuiController{bi: bi}
}

// Exec コマンド実行
func (c *BidEuchreCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.bi.GetConfig()
			return c.bi.ResetWithConfig(cfg)
		},
		[]string{"b", "bid", "ps", "pass", "t", "trump", "p", "play", "n", "next", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				return cuiutil.WithParsedInt(args, "Bid value is required.", "Invalid bid: %s.",
					domain.BidEuchreMinBid, domain.BidEuchreMaxBid, func(v int) string {
						return c.bi.Bid(v)
					})
			case "ps", "pass":
				return c.bi.PassBid(), true
			case "t", "trump":
				// **0=S 1=C 2=D 3=H 4=NT-high 5=NT-low。**ノートランプが 2 種類ある。
				return cuiutil.WithParsedInt(args, "Trump declaration is required (0=S 1=C 2=D 3=H 4=NT-high 5=NT-low).",
					"Invalid trump: %s.", 0, int(domain.BidEuchreTrumpCount)-1, func(v int) string {
						return c.bi.ChooseTrump(v)
					})
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", 0, domain.BidEuchreHandSize-1, func(v int) string {
					return c.bi.PlayCard(v)
				})
			case "n", "next":
				return c.bi.NextHand(), true
			default:
				return handleCuiLog(cmd, c.bi.ActionLog)
			}
		},
	)
}
