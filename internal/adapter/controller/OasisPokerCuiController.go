//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OasisPokerCuiController オアシスポーカーCUIコントローラークラス
type OasisPokerCuiController struct {
	oi usecase.OasisPokerInteractorIF
}

// NewOasisPokerCuiController コンストラクタ
func NewOasisPokerCuiController(oi usecase.OasisPokerInteractorIF) *OasisPokerCuiController {
	return &OasisPokerCuiController{
		oi: oi,
	}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "b 100 10", "e 0 2 4", "s", "p", "f", "q"
func (oc *OasisPokerCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return oc.oi.Reset() },
		[]string{"b", "bet", "e", "exchange", "s", "stand", "p", "play", "f", "fold", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				ante, errMsg, ok := cuiutil.ParseIntArgKeys(args, "anteAmountRequired", "invalidAnteAmount", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				jackpot := 0
				if len(args) > 1 {
					jackpot, errMsg, ok = cuiutil.ParseIntArg(args[1:], "", "Invalid jackpot amount.", 0, math.MaxInt)
					if !ok {
						return errMsg, true
					}
				}
				return oc.oi.Bet(ante, jackpot), true
			case "e", "exchange":
				indices, skipped := cuiutil.ParseBoundedIntSlice(args, 0, 4)
				return cuiutil.PrependSkippedWarning(oc.oi.Exchange(indices), skipped), true
			case "s", "stand":
				return oc.oi.Stand(), true
			case "p", "play":
				return oc.oi.Play(), true
			case "f", "fold":
				return oc.oi.Fold(), true
			default:
				return handleCuiHintAndLog(cmd, oc.oi.Hint, oc.oi.ActionLog)
			}
		},
	)
}
