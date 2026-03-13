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
// コマンド例: "r", "b 100 0", "q"
func (bcc *BaccaratCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return bcc.bi.Reset() },
		unknownCommandMessage,
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				if len(args) < 2 {
					return "Usage: b <amount> <betType(0=Player,1=Banker,2=Tie)>", true
				}
				amount, errMsg, ok := cuiutil.ParseIntArg(args, "", "Invalid bet amount. Please enter a number.", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				betType, errMsg2, ok2 := cuiutil.ParseIntArg(args[1:], "", "Invalid bet type. Please enter 0(Player), 1(Banker), or 2(Tie).", 0, 2)
				if !ok2 {
					return errMsg2, true
				}
				return bcc.bi.Bet(amount, betType), true
			case "log", "l":
				return bcc.bi.ActionLog(), true
			}
			return "", false
		},
	)
}
