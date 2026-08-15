//go:build !js || !wasm || extra3

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BostonCuiController ボストン (Boston) CUIコントローラークラス
type BostonCuiController struct {
	bi usecase.BostonInteractorIF
}

// NewBostonCuiController コンストラクタ
func NewBostonCuiController(bi usecase.BostonInteractorIF) *BostonCuiController {
	return &BostonCuiController{bi: bi}
}

// Exec コマンド実行
func (c *BostonCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.bi.GetConfig()
			return c.bi.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bid", "ps", "pass", "cp", "callpartner",
			"p", "play", "n", "next", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				return bostonParseBid(args, c.bi)
			case "ps", "pass":
				return c.bi.PassBid(), true
			case "cp", "callpartner":
				return bostonParseCallPartner(args, c.bi)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", 0, domain.BostonHandSize-1, func(v int) string {
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

// bostonParseBid は `b <level> [suit]` を解釈する。
//
// **序列はトリック数ではなく段の番号で指す。**ミゼールが間に挟まるので、
// トリック数だけでは一意に決まらない。
func bostonParseBid(args []string, bi usecase.BostonInteractorIF) (string, bool) {
	if len(args) == 0 {
		return invalidArg("bidLevelRequiredLadder"), true
	}
	level, err := strconv.Atoi(args[0])
	if err != nil || level <= int(domain.BostonBidPass) || level >= int(domain.BostonBidLevelCount) {
		return "Invalid bid level: " + args[0] + ". Please enter 1-" +
			strconv.Itoa(int(domain.BostonBidLevelCount)-1) + ".", true
	}
	suit := 0
	if len(args) > 1 {
		suit, err = strconv.Atoi(args[1])
		if err != nil || suit < domain.CardDesignSpade || suit > domain.CardDesignDiamond {
			return invalidArg("invalidSuit14Plain", "val", args[1]), true
		}
	}
	return bi.Bid(domain.BostonBidLevel(level), suit), true
}

// bostonParseCallPartner は `cp <seat|-1>` を解釈する。
func bostonParseCallPartner(args []string, bi usecase.BostonInteractorIF) (string, bool) {
	if len(args) == 0 {
		return invalidArg("partnerSeatRequired"), true
	}
	seat, err := strconv.Atoi(args[0])
	if err != nil || seat < -1 || seat >= domain.BostonPlayerCnt {
		return "Invalid partner: " + args[0] + ". Please enter -1 to " +
			strconv.Itoa(domain.BostonPlayerCnt-1) + ".", true
	}
	return bi.CallPartner(seat), true
}
