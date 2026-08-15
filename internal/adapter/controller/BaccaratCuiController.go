//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BaccaratCuiController バカラCUIコントローラークラス
type BaccaratCuiController struct {
	bi usecase.BaccaratInteractorIF
}

// NewBaccaratCuiController コンストラクタ
func NewBaccaratCuiController(bi usecase.BaccaratInteractorIF) *BaccaratCuiController {
	return &BaccaratCuiController{
		bi: bi,
	}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100 0", "ch", "q"
func (bcc *BaccaratCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return bcc.bi.Reset() },
		[]string{"b", "bet", "log", "l", "ch", "clearhistory"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args, "betAmountRequired", "invalidBetAmount", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				betType, errMsg, ok := cuiutil.ParseIntArgKeys(args[1:], "betTypeRequired0Player1Banker2Tie", "invalidBetType0Player1BankerOr2Tie", 0, 2)
				if !ok {
					return errMsg, true
				}
				return bcc.bi.Bet(amount, betType, 0, 0), true
			case "ch", "clearhistory":
				return bcc.bi.ClearHistory(), true
			default:
				return handleCuiLog(cmd, bcc.bi.ActionLog)
			}
		},
	)
}
