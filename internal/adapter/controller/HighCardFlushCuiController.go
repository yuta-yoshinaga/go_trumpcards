//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HighCardFlushCuiController ハイカードフラッシュCUIコントローラークラス
type HighCardFlushCuiController struct {
	hi usecase.HighCardFlushInteractorIF
}

// NewHighCardFlushCuiController コンストラクタ
func NewHighCardFlushCuiController(hi usecase.HighCardFlushInteractorIF) *HighCardFlushCuiController {
	return &HighCardFlushCuiController{
		hi: hi,
	}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "b 100 50 20", "ra 1", "f", "q", "log"
func (hcc *HighCardFlushCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return hcc.hi.Reset() },
		[]string{"b", "bet", "ra", "raise", "f", "fold", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				ante, errMsg, ok := cuiutil.ParseIntArg(args, "Ante amount is required.", "Invalid ante amount. Please enter a number.", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				flushBonus := 0
				if len(args) > 1 {
					flushBonus, errMsg, ok = cuiutil.ParseIntArg(args[1:], "", "Invalid Flush Bonus amount.", 0, math.MaxInt)
					if !ok {
						return errMsg, true
					}
				}
				straightFlush := 0
				if len(args) > 2 {
					straightFlush, errMsg, ok = cuiutil.ParseIntArg(args[2:], "", "Invalid Straight Flush Bonus amount.", 0, math.MaxInt)
					if !ok {
						return errMsg, true
					}
				}
				return hcc.hi.Bet(ante, flushBonus, straightFlush), true
			case "ra", "raise":
				mult, errMsg, ok := cuiutil.ParseIntArg(args, "Raise multiplier is required (1-3).", "Invalid raise multiplier.", 1, 3)
				if !ok {
					return errMsg, true
				}
				return hcc.hi.Raise(mult), true
			case "f", "fold":
				return hcc.hi.Fold(), true
			default:
				return handleCuiHintAndLog(cmd, hcc.hi.Hint, hcc.hi.ActionLog)
			}
		},
	)
}
