//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CrazyFourPokerCuiController クレイジー 4 ポーカーCUIコントローラークラス
type CrazyFourPokerCuiController struct {
	ci usecase.CrazyFourPokerInteractorIF
}

// NewCrazyFourPokerCuiController コンストラクタ
func NewCrazyFourPokerCuiController(ci usecase.CrazyFourPokerInteractorIF) *CrazyFourPokerCuiController {
	return &CrazyFourPokerCuiController{ci: ci}
}

// Exec ゲーム実行
// コマンド例: "r", "bet 50 20", "play 3", "fold", "next", "hint", "log", "q"
//   - bet の 2 つ目 (Queens Up) は省略できる
//   - play の倍率は 1〜3。**3 はエースのペア以上でのみ通る** (ドメインが判定する)
func (cc *CrazyFourPokerCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.ci.Reset() },
		[]string{"bet", "b", "play", "p", "fold", "f", "next", "hint", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				ante, errMsg, ok := cuiutil.ParseIntArg(args,
					"Ante is required.", "Invalid ante. Please enter a number.", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				// **Queens Up は省略可。** 省略は「置かない」。
				queensUp := 0
				if len(args) > 1 {
					v, errMsg2, ok2 := cuiutil.ParseIntArg(args[1:],
						"", "Invalid queens up bet. Please enter a number.", 0, math.MaxInt)
					if !ok2 {
						return errMsg2, true
					}
					queensUp = v
				}
				return cc.ci.PlaceBet(ante, queensUp), true
			case "p", "play":
				// **上限は書かない。** 手役次第で 1 か 3 に変わるのでドメインが弾く。
				mult, errMsg, ok := cuiutil.ParseIntArg(args,
					"Play multiplier is required.", "Invalid multiplier. Please enter a number.", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return cc.ci.Play(mult), true
			case "f", "fold":
				return cc.ci.Fold(), true
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
