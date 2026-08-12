//go:build !js || !wasm || classic

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ColourWhistCuiController カラーホイストCUIコントローラークラス
type ColourWhistCuiController struct {
	ci usecase.ColourWhistInteractorIF
}

// NewColourWhistCuiController コンストラクタ
func NewColourWhistCuiController(ci usecase.ColourWhistInteractorIF) *ColourWhistCuiController {
	return &ColourWhistCuiController{ci: ci}
}

// Exec ゲーム実行
// コマンド例: "r", "bid samen", "pass", "call h", "p 3", "next", "hint", "log", "q"
//   - 契約: samen / alleen / miserie（**troel は配りでしか成立しないので指定できません**）
//   - 切り札: s/c/h/d
func (cc *ColourWhistCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.ci.Reset() },
		[]string{"p", "play", "bid", "pass", "call", "next", "giveup", "hint", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				idx, errMsg, ok := cuiutil.ParseIntArg(args,
					"Card index is required.", "Invalid card index. Please enter a number.", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return cc.ci.PlayCard(idx), true
			case "bid":
				contract, ok := colourWhistParseContract(args)
				if !ok {
					return "Invalid contract. Use samen / alleen / miserie, or pass (troel is dealt, not bid).", true
				}
				return cc.ci.Bid(contract), true
			case "pass":
				return cc.ci.Bid(domain.ColourWhistContractNone), true
			case "call":
				suit, ok := colourWhistParseSuit(args)
				if !ok {
					return "Invalid trump. Use s / c / h / d.", true
				}
				return cc.ci.Call(suit), true
			case "next":
				return cc.ci.NextRound(), true
			case "giveup":
				return cc.ci.GiveUp(), true
			case "hint":
				return cc.ci.Hint(), true
			default:
				return handleCuiLog(cmd, cc.ci.ActionLog)
			}
		},
	)
}

// colourWhistParseContract は契約を文字列から解析する。
//
// **troel は受け付けません。** 配りでしか成立しない契約なので、競りの語彙には
// 入れないほうが誤解がありません（送られてもドメインが弾きます）。
func colourWhistParseContract(args []string) (int, bool) {
	if len(args) == 0 {
		return 0, false
	}
	switch args[0] {
	case "pass":
		return domain.ColourWhistContractNone, true
	case "samen":
		return domain.ColourWhistContractSamen, true
	case "alleen":
		return domain.ColourWhistContractAlleen, true
	case "miserie", "mis":
		return domain.ColourWhistContractMiserie, true
	default:
		return 0, false
	}
}

// colourWhistParseSuit は切り札を文字列から解析する。
func colourWhistParseSuit(args []string) (int, bool) {
	if len(args) == 0 {
		return 0, false
	}
	switch args[0] {
	case "s", "spade":
		return domain.CardDesignSpade, true
	case "c", "clover", "club":
		return domain.CardDesignClover, true
	case "h", "heart":
		return domain.CardDesignHeart, true
	case "d", "diamond":
		return domain.CardDesignDiamond, true
	default:
		return 0, false
	}
}
