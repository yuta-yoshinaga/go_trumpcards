//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CasinoHoldemCuiController カジノホールデムCUIコントローラークラス
type CasinoHoldemCuiController struct {
	ci usecase.CasinoHoldemInteractorIF
}

// NewCasinoHoldemCuiController コンストラクタ
func NewCasinoHoldemCuiController(ci usecase.CasinoHoldemInteractorIF) *CasinoHoldemCuiController {
	return &CasinoHoldemCuiController{ci: ci}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "b 100 10", "c", "f", "log"
func (cc *CasinoHoldemCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.ci.Reset() },
		[]string{"b", "bet", "c", "call", "f", "fold", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				ante, errMsg, ok := cuiutil.ParseIntArgKeys(args, "anteAmountRequired", "invalidAnteAmount", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				bonus := 0
				if len(args) > 1 {
					bonus, errMsg, ok = cuiutil.ParseIntArgKeys(args[1:], "", "invalidBonusAmount", 0, math.MaxInt)
					if !ok {
						return errMsg, true
					}
				}
				return cc.ci.Bet(ante, bonus), true
			case "c", "call":
				return cc.ci.Call(), true
			case "f", "fold":
				return cc.ci.Fold(), true
			default:
				return handleCuiHintAndLog(cmd, cc.ci.Hint, cc.ci.ActionLog)
			}
		},
	)
}
