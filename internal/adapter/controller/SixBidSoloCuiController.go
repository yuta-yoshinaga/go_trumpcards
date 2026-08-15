//go:build !js || !wasm || extra2

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SixBidSoloCuiController シックスビッド・ソロ (Six-Bid Solo) CUIコントローラークラス
type SixBidSoloCuiController struct {
	si usecase.SixBidSoloInteractorIF
}

// NewSixBidSoloCuiController コンストラクタ
func NewSixBidSoloCuiController(si usecase.SixBidSoloInteractorIF) *SixBidSoloCuiController {
	return &SixBidSoloCuiController{si: si}
}

// Exec コマンド実行
func (c *SixBidSoloCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.si.GetConfig()
			return c.si.ResetWithConfig(cfg)
		},
		[]string{"b", "bid", "ps", "pass", "d", "declare", "p", "play", "n", "next", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				// **1=ソロ 2=ハートソロ 3=ミゼール 4=ギャランティー 5=スプレッド 6=コール。**
				return cuiutil.WithParsedIntKeys(args, "bidRequiredSolo", "invalidBid", int(domain.SixBidSoloMinBid), int(domain.SixBidSoloMaxBid), func(v int) string {
					return c.si.Bid(v)
				})
			case "ps", "pass":
				return c.si.PassBid(), true
			case "d", "declare":
				return sixBidSoloParseDeclare(args, c.si)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
					0, domain.SixBidSoloHandSize-1, func(v int) string {
						return c.si.PlayCard(v)
					})
			case "n", "next":
				return c.si.NextHand(), true
			default:
				return handleCuiLog(cmd, c.si.ActionLog)
			}
		},
	)
}

// sixBidSoloParseDeclare は `d <suit> [calledSuit calledValue]` を解釈する。
//
// **指名札はコール・ソロのときだけ要る。**スートは 1=♠ 2=♣ 3=♥ 4=♦。
func sixBidSoloParseDeclare(args []string, si usecase.SixBidSoloInteractorIF) (string, bool) {
	if len(args) < 1 {
		return invalidArg("trumpSuitRequiredLettersPlainRaw"), true
	}
	suit, err := strconv.Atoi(args[0])
	if err != nil || suit < domain.CardDesignSpade || suit > domain.CardDesignDiamond {
		return invalidArg("invalidSuit14Letters", "val", args[0]), true
	}
	if len(args) == 1 {
		return si.Declare(suit, 0, 0), true
	}
	if len(args) < 3 {
		return "A call solo needs both the called suit and its value (e.g. d 1 1 1).", true
	}
	calledSuit, err := strconv.Atoi(args[1])
	if err != nil || calledSuit < domain.CardDesignSpade || calledSuit > domain.CardDesignDiamond {
		return invalidArg("invalidCalledSuit14", "val", args[1]), true
	}
	calledValue, err := strconv.Atoi(args[2])
	if err != nil || calledValue < 1 || calledValue > 13 {
		return invalidArg("invalidCalledValue113", "val", args[2]), true
	}
	return si.Declare(suit, calledSuit, calledValue), true
}
