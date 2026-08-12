//go:build !js || !wasm || classic

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BotifarraCuiController ボティファラCUIコントローラークラス
type BotifarraCuiController struct {
	bi usecase.BotifarraInteractorIF
}

// NewBotifarraCuiController コンストラクタ
func NewBotifarraCuiController(bi usecase.BotifarraInteractorIF) *BotifarraCuiController {
	return &BotifarraCuiController{bi: bi}
}

// Exec ゲーム実行
// コマンド例: "r", "declare s", "delegate", "double", "pass", "p 3", "next", "hint", "log", "q"
//   - 切り札: s/spade, c/clover, h/heart, d/diamond, n/none (切り札なし)
func (bc *BotifarraCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return bc.bi.Reset() },
		[]string{"p", "play", "declare", "delegate", "double", "pass", "next", "giveup", "hint", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				idx, errMsg, ok := cuiutil.ParseIntArg(args,
					"Card index is required.", "Invalid card index. Please enter a number.", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return bc.bi.PlayCard(idx), true
			case "declare":
				suit, ok := botifarraParseTrump(args)
				if !ok {
					return "Invalid trump. Use s/c/h/d, or n for no trump.", true
				}
				return bc.bi.Declare(suit), true
			case "delegate":
				return bc.bi.Delegate(), true
			case "double":
				return bc.bi.Double(), true
			case "pass":
				return bc.bi.PassDouble(), true
			case "next":
				return bc.bi.NextRound(), true
			case "giveup":
				return bc.bi.GiveUp(), true
			case "hint":
				return bc.bi.Hint(), true
			default:
				return handleCuiLog(cmd, bc.bi.ActionLog)
			}
		},
	)
}

// botifarraParseTrump は切り札を文字列から解析する。
//
// **切り札なし (n) も有効な宣言**なので、「指定が無い」とは区別します。
func botifarraParseTrump(args []string) (int, bool) {
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
	case "n", "none", "notrump":
		return domain.BotifarraNoTrump, true
	default:
		return 0, false
	}
}
