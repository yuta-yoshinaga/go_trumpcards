//go:build !js || !wasm || extra2

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RikkenCuiController リッケンCUIコントローラークラス
type RikkenCuiController struct {
	ri usecase.RikkenInteractorIF
}

// NewRikkenCuiController コンストラクタ
func NewRikkenCuiController(ri usecase.RikkenInteractorIF) *RikkenCuiController {
	return &RikkenCuiController{ri: ri}
}

// Exec ゲーム実行
// コマンド例: "r", "bid rik", "pass", "call h", "p 3", "next", "hint", "log", "q"
//   - 契約: rik / misere / solo / open (open misere), pass
//   - 切り札: s/c/h/d
func (rc *RikkenCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return rc.ri.Reset() },
		[]string{"p", "play", "bid", "pass", "call", "next", "giveup", "hint", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				idx, errMsg, ok := cuiutil.ParseIntArgKeys(args,
					"cardIndexRequired", "invalidCardIndexNotANumber", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return rc.ri.PlayCard(idx), true
			case "bid":
				contract, ok := rikkenParseContract(args)
				if !ok {
					return invalidArg("invalidContractRik"), true
				}
				return rc.ri.Bid(contract), true
			case "pass":
				// **パスは契約 0。** 別の経路にはしません。
				return rc.ri.Bid(domain.RikkenContractNone), true
			case "call":
				suit, ok := rikkenParseSuit(args)
				if !ok {
					return invalidArg("invalidTrumpSCHD"), true
				}
				return rc.ri.Call(suit), true
			case "next":
				return rc.ri.NextRound(), true
			case "giveup":
				return rc.ri.GiveUp(), true
			case "hint":
				return rc.ri.Hint(), true
			default:
				return handleCuiLog(cmd, rc.ri.ActionLog)
			}
		},
	)
}

// rikkenParseContract は契約を文字列から解析する。
func rikkenParseContract(args []string) (int, bool) {
	if len(args) == 0 {
		return 0, false
	}
	switch args[0] {
	case "pass":
		return domain.RikkenContractNone, true
	case "rik":
		return domain.RikkenContractRik, true
	case "misere", "mis":
		return domain.RikkenContractMisere, true
	case "solo":
		return domain.RikkenContractSolo, true
	case "open", "openmisere":
		return domain.RikkenContractOpenMisere, true
	default:
		return 0, false
	}
}

// rikkenParseSuit は切り札を文字列から解析する。
func rikkenParseSuit(args []string) (int, bool) {
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
