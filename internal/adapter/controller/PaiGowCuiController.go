//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PaiGowCuiController パイガオポーカーCUIコントローラークラス
type PaiGowCuiController struct {
	pi usecase.PaiGowInteractorIF
}

// NewPaiGowCuiController コンストラクタ
func NewPaiGowCuiController(pi usecase.PaiGowInteractorIF) *PaiGowCuiController {
	return &PaiGowCuiController{
		pi: pi,
	}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "s 0 1", "q"
func (pgc *PaiGowCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return pgc.pi.Reset() },
		[]string{"b", "bet", "s", "set", "a", "auto", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args, "betAmountRequired", "invalidBetAmount", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return pgc.pi.Bet(amount), true
			case "s", "set":
				low0, errMsg, ok := cuiutil.ParseIntArgKeys(args, "twoCardIndicesAreRequired", "invalidFirstIndex", 0, 6)
				if !ok {
					return errMsg, true
				}
				if len(args) < 2 {
					return "Two card indices are required.", true
				}
				low1, errMsg, ok := cuiutil.ParseIntArgKeys(args[1:], "secondIndexRequired", "invalidSecondIndex", 0, 6)
				if !ok {
					return errMsg, true
				}
				return pgc.pi.SetHands(low0, low1), true
			case "a", "auto":
				return pgc.pi.AutoSetHands(), true
			case "h", "hint":
				return pgc.pi.Hint(), true
			default:
				return handleCuiLog(cmd, pgc.pi.ActionLog)
			}
		},
	)
}
