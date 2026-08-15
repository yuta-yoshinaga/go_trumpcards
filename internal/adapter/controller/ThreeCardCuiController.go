//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ThreeCardCuiController スリーカードポーカーCUIコントローラークラス
type ThreeCardCuiController struct {
	ti usecase.ThreeCardInteractorIF
}

// NewThreeCardCuiController コンストラクタ
func NewThreeCardCuiController(ti usecase.ThreeCardInteractorIF) *ThreeCardCuiController {
	return &ThreeCardCuiController{
		ti: ti,
	}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "b 100 50", "p", "f", "q"
func (tcc *ThreeCardCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return tcc.ti.Reset() },
		[]string{"b", "bet", "p", "play", "f", "fold", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				ante, errMsg, ok := cuiutil.ParseIntArgKeys(args, "anteAmountRequired", "invalidAnteAmount", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				pairPlus := 0
				if len(args) > 1 {
					pairPlus, errMsg, ok = cuiutil.ParseIntArg(args[1:], "", "Invalid Pair Plus amount.", 0, math.MaxInt)
					if !ok {
						return errMsg, true
					}
				}
				return tcc.ti.Bet(ante, pairPlus), true
			case "p", "play":
				return tcc.ti.Play(), true
			case "f", "fold":
				return tcc.ti.Fold(), true
			default:
				return handleCuiHintAndLog(cmd, tcc.ti.Hint, tcc.ti.ActionLog)
			}
		},
	)
}
