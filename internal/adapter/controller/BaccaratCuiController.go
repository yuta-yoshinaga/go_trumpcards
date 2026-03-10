package controller

import (
	"strconv"

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
				amount, err := strconv.Atoi(args[0])
				if err != nil || amount <= 0 {
					return "Invalid bet amount. Please enter a number.", true
				}
				betType, err := strconv.Atoi(args[1])
				if err != nil || betType < 0 || betType > 2 {
					return "Invalid bet type. Please enter 0(Player), 1(Banker), or 2(Tie).", true
				}
				return bcc.bi.Bet(amount, betType), true
			case "log", "l":
				return bcc.bi.ActionLog(), true
			}
			return "", false
		},
	)
}
