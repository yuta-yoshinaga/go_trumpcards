//go:build !js || !wasm || extra3

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// VintCuiController ヴィント (Vint) CUIコントローラークラス
type VintCuiController struct {
	vi usecase.VintInteractorIF
}

// NewVintCuiController コンストラクタ
func NewVintCuiController(vi usecase.VintInteractorIF) *VintCuiController {
	return &VintCuiController{vi: vi}
}

// Exec コマンド実行
func (c *VintCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.vi.GetConfig()
			return c.vi.ResetWithConfig(cfg)
		},
		[]string{"b", "bid", "ps", "pass", "p", "play", "n", "next", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				return vintParseBid(args, c.vi)
			case "ps", "pass":
				return c.vi.PassBid(), true
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", 0, domain.VintHandSize-1, func(v int) string {
					return c.vi.PlayCard(v)
				})
			case "n", "next":
				return c.vi.NextHand(), true
			default:
				return handleCuiLog(cmd, c.vi.ActionLog)
			}
		},
	)
}

// vintParseBid は `b <level> <denom>` を解釈する。
//
// **denom は 0=♠ 1=♣ 2=♦ 3=♥ 4=NT。**ブリッジとは序列が違うので番号で指す。
func vintParseBid(args []string, vi usecase.VintInteractorIF) (string, bool) {
	if len(args) < 2 {
		return invalidArg("bidLevelAndDenominationRequired"), true
	}
	level, err := strconv.Atoi(args[0])
	if err != nil || level < domain.VintMinLevel || level > domain.VintMaxLevel {
		return invalidArg("invalidBidLevelMinMax", "val", args[0], "max", strconv.Itoa(domain.VintMinLevel)+"-"+strconv.Itoa(domain.VintMaxLevel)), true
	}
	denom, err := strconv.Atoi(args[1])
	if err != nil || denom < 0 || denom >= domain.VintDenomCount {
		return invalidArg("invalidDenomination0Max", "val", args[1], "max", strconv.Itoa(domain.VintDenomCount-1)), true
	}
	return vi.Bid(level, denom), true
}
