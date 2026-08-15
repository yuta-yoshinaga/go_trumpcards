//go:build !js || !wasm || extra3

package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KaiserCuiController カイザー (Kaiser) CUIコントローラークラス
type KaiserCuiController struct {
	ki usecase.KaiserInteractorIF
}

// NewKaiserCuiController コンストラクタ
func NewKaiserCuiController(ki usecase.KaiserInteractorIF) *KaiserCuiController {
	return &KaiserCuiController{ki: ki}
}

// Exec コマンド実行
func (c *KaiserCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ki.GetConfig()
			return c.ki.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bid", "ps", "pass", "t", "trump",
			"d", "discard", "p", "play", "n", "next", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				return kaiserParseBid(args, c.ki)
			case "ps", "pass":
				return c.ki.PassBid(), true
			case "t", "trump":
				return cuiutil.WithParsedIntKeys(args, "suitRequiredLetters", "invalidSuitRange", domain.CardDesignSpade, domain.CardDesignDiamond, func(v int) string {
					return c.ki.SetTrump(v)
				})
			case "d", "discard":
				return kaiserParseDiscard(args, c.ki)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", 0, domain.KaiserHandSize+domain.KaiserKittySize-1, func(v int) string {
					return c.ki.PlayCard(v)
				})
			case "n", "next":
				return c.ki.NextHand(), true
			default:
				return handleCuiHintAndLog(cmd, c.ki.Hint, c.ki.ActionLog)
			}
		},
	)
}

// kaiserParseBid は `b <7-12> [0-2]` を解釈する。
//
// 第 2 引数は契約種別 (0=切札, 1=ノートランプ, 2=ロー・ノートランプ)。省略時は切札。
func kaiserParseBid(args []string, ki usecase.KaiserInteractorIF) (string, bool) {
	if len(args) == 0 {
		return invalidArg("bidValueRequired712"), true
	}
	value, err := strconv.Atoi(args[0])
	if err != nil || value < domain.KaiserMinBid || value > domain.KaiserMaxBid {
		return "Invalid bid: " + args[0] + ". Please enter " +
			strconv.Itoa(domain.KaiserMinBid) + "-" + strconv.Itoa(domain.KaiserMaxBid) + ".", true
	}
	contract := 0
	if len(args) > 1 {
		contract, err = strconv.Atoi(args[1])
		if err != nil || contract < int(domain.KaiserContractTrump) || contract > int(domain.KaiserContractLowNoTrump) {
			return invalidArg("invalidContract02", "val", args[1]), true
		}
	}
	return ki.Bid(value, domain.KaiserContract(contract)), true
}

// kaiserParseDiscard は `d <i> <j>` を解釈する。
func kaiserParseDiscard(args []string, ki usecase.KaiserInteractorIF) (string, bool) {
	if len(args) < domain.KaiserKittySize {
		return invalidArg("twoIndicesRequired"), true
	}
	idxs := make([]int, 0, domain.KaiserKittySize)
	for _, a := range args[:domain.KaiserKittySize] {
		v, err := strconv.Atoi(strings.TrimSpace(a))
		if err != nil || v < 0 {
			return invalidArg("invalidCardIndex", "val", a), true
		}
		idxs = append(idxs, v)
	}
	return ki.Discard(idxs), true
}
