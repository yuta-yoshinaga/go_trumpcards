//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TexasHoldemBonusCuiController テキサスホールデムボーナスポーカーCUIコントローラークラス
type TexasHoldemBonusCuiController struct {
	ti usecase.TexasHoldemBonusInteractorIF
}

// NewTexasHoldemBonusCuiController コンストラクタ
func NewTexasHoldemBonusCuiController(ti usecase.TexasHoldemBonusInteractorIF) *TexasHoldemBonusCuiController {
	return &TexasHoldemBonusCuiController{ti: ti}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "b 100 10", "p", "f", "c", "ra", "log"
func (tc *TexasHoldemBonusCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return tc.ti.Reset() },
		[]string{"b", "bet", "p", "play", "f", "fold", "c", "check", "ra", "raise", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				ante, errMsg, ok := cuiutil.ParseIntArgKeys(args, "anteAmountRequired", "invalidAnteAmount", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				bonus := 0
				if len(args) > 1 {
					bonus, errMsg, ok = cuiutil.ParseIntArgKeys(args[1:], "", "invalidBonusAmount", 0, math.MaxInt)
					if !ok {
						return errMsg, true
					}
				}
				return tc.ti.Bet(ante, bonus), true
			case "p", "play":
				return tc.ti.Play(), true
			case "f", "fold":
				return tc.ti.Fold(), true
			case "c", "check":
				return tc.ti.Check(), true
			case "ra", "raise":
				return tc.ti.Raise(), true
			default:
				return handleCuiLog(cmd, tc.ti.ActionLog)
			}
		},
	)
}
