//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MississippiStudCuiController ミシシッピ・スタッドCUIコントローラー
type MississippiStudCuiController struct {
	ci usecase.MississippiStudInteractorIF
}

// NewMississippiStudCuiController コンストラクタ
func NewMississippiStudCuiController(ci usecase.MississippiStudInteractorIF) *MississippiStudCuiController {
	return &MississippiStudCuiController{ci: ci}
}

// Exec ゲーム実行。
// コマンド例: "r", "b 100", "p 1", "p 2", "p 3", "f", "log", "q"
func (mc *MississippiStudCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return mc.ci.Reset() },
		[]string{"b", "bet", "p", "play", "f", "fold", "h", "hint", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args, "betAmountRequired", "invalidBetAmount", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return mc.ci.Bet(amount), true
			case "p", "play":
				mult, errMsg, ok := cuiutil.ParseIntArg(args, "Multiplier (1, 2 or 3) is required.", "Invalid multiplier. Please enter a number.", 1, 3)
				if !ok {
					return errMsg, true
				}
				return mc.ci.Play(mult), true
			case "f", "fold":
				return mc.ci.Fold(), true
			case "h", "hint":
				return mc.ci.Hint(), true
			default:
				return handleCuiLog(cmd, mc.ci.ActionLog)
			}
		},
	)
}
