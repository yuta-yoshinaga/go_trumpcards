package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// LetItRideCuiController レット・イット・ライドCUIコントローラークラス
type LetItRideCuiController struct {
	ci usecase.LetItRideInteractorIF
}

// NewLetItRideCuiController コンストラクタ
func NewLetItRideCuiController(ci usecase.LetItRideInteractorIF) *LetItRideCuiController {
	return &LetItRideCuiController{
		ci: ci,
	}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "p", "l", "log", "q"
func (lc *LetItRideCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return lc.ci.Reset() },
		[]string{"b", "bet", "p", "pull", "l", "letitride", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				amount, errMsg, ok := cuiutil.ParseIntArg(args, "Bet amount is required.", "Invalid bet amount. Please enter a number.", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return lc.ci.Bet(amount), true
			case "p", "pull":
				return lc.ci.Pull(), true
			case "l", "letitride":
				return lc.ci.LetItRide(), true
			default:
				return handleCuiLog(cmd, lc.ci.ActionLog)
			}
		},
	)
}
