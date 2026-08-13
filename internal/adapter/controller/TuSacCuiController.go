//go:build !js || !wasm || solo

package controller

import (
	"math"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TuSacCuiController 四色牌CUIコントローラークラス
type TuSacCuiController struct {
	ci usecase.TuSacInteractorIF
}

// NewTuSacCuiController コンストラクタ
func NewTuSacCuiController(ci usecase.TuSacInteractorIF) *TuSacCuiController {
	return &TuSacCuiController{ci: ci}
}

// Exec ゲーム実行
// コマンド例: "r", "draw", "take", "meld 1 4 7", "discard 3", "next", "hint", "log", "q"
func (cc *TuSacCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.ci.Reset() },
		[]string{"draw", "d", "take", "t", "meld", "m", "discard", "x", "next", "hint", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			// **山と捨て札は別のコマンド。** 引き先を引数にすると、付け忘れが
			// 「山から」に化けて、狙って拾った札が黙って流れる。
			case "draw", "d":
				return cc.ci.Draw(false), true
			case "take", "t":
				return cc.ci.Draw(true), true
			case "meld", "m":
				indexes, errMsg, ok := tuSacParseIndexes(args)
				if !ok {
					return errMsg, true
				}
				return cc.ci.Meld(indexes), true
			case "discard", "x":
				index, errMsg, ok := cuiutil.ParseIntArg(args,
					"Card index is required.", "Invalid index. Please enter a number.", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				// **画面は 1 始まりで見せる。** 内部は 0 始まり。
				return cc.ci.Discard(index - 1), true
			case "next":
				return cc.ci.NextRound(), true
			case "hint":
				return cc.ci.Hint(), true
			default:
				return handleCuiLog(cmd, cc.ci.ActionLog)
			}
		},
	)
}

// tuSacParseIndexes は "1 4 7" のような並びを 0 始まりの添字にする。
func tuSacParseIndexes(args []string) ([]int, string, bool) {
	if len(args) == 0 {
		return nil, "Card indexes are required (e.g. meld 1 4 7).", false
	}
	out := make([]int, 0, len(args))
	for _, a := range args {
		n, err := strconv.Atoi(strings.TrimSpace(a))
		if err != nil || n < 1 {
			return nil, "Invalid index. Please enter numbers from 1.", false
		}
		out = append(out, n-1)
	}
	return out, "", true
}
