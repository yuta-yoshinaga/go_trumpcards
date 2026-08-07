//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// UltimateTexasHoldemCuiController アルティメット・テキサスホールデムCUIコントローラークラス
type UltimateTexasHoldemCuiController struct {
	ui usecase.UltimateTexasHoldemInteractorIF
}

// NewUltimateTexasHoldemCuiController コンストラクタ
func NewUltimateTexasHoldemCuiController(ui usecase.UltimateTexasHoldemInteractorIF) *UltimateTexasHoldemCuiController {
	return &UltimateTexasHoldemCuiController{ui: ui}
}

// Exec ゲーム実行
//
// コマンド例: "r", "b 100", "b 100 10", "p 4", "p 3", "p 2", "p 1", "c", "f", "log"
func (uc *UltimateTexasHoldemCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return uc.ui.Reset() },
		[]string{"b", "bet", "p", "play", "c", "check", "f", "fold", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				ante, errMsg, ok := cuiutil.ParseIntArg(args, "Ante amount is required.", "Invalid ante amount. Please enter a number.", domain.UltimateTexasHoldemMinBet, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				trips := 0
				if len(args) > 1 {
					trips, errMsg, ok = cuiutil.ParseIntArg(args[1:], "", "Invalid trips amount.", 0, math.MaxInt)
					if !ok {
						return errMsg, true
					}
				}
				return uc.ui.Bet(ante, trips), true
			case "p", "play":
				mult, errMsg, ok := cuiutil.ParseIntArg(args, "Play multiplier is required (e.g. `p 4`).", "Invalid play multiplier.", 1, 4)
				if !ok {
					return errMsg, true
				}
				return uc.ui.Play(mult), true
			case "c", "check":
				return uc.ui.Check(), true
			case "f", "fold":
				return uc.ui.Fold(), true
			case "h", "hint":
				return uc.ui.Hint(), true
			default:
				return handleCuiLog(cmd, uc.ui.ActionLog)
			}
		},
	)
}
