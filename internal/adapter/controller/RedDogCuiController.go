//go:build !js || !wasm || extra4

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RedDogCuiController レッドドッグCUIコントローラークラス
type RedDogCuiController struct {
	ci usecase.RedDogInteractorIF
}

// NewRedDogCuiController コンストラクタ
func NewRedDogCuiController(ci usecase.RedDogInteractorIF) *RedDogCuiController {
	return &RedDogCuiController{ci: ci}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "raise 50", "s", "log", "q"
func (rc *RedDogCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return rc.ci.Reset() },
		[]string{"b", "bet", "raise", "s", "stay", "h", "hint", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args, "betAmountRequired", "invalidBetAmount", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return rc.ci.Bet(amount), true
			case "raise":
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args, "raiseAmountRequired", "invalidRaiseAmountANumber", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return rc.ci.Raise(amount), true
			case "s", "stay":
				return rc.ci.Stay(), true
			case "h", "hint":
				return rc.ci.Hint(), true
			default:
				return handleCuiLog(cmd, rc.ci.ActionLog)
			}
		},
	)
}
