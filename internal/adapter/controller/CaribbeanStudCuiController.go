//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CaribbeanStudCuiController カリビアンスタッドポーカーCUIコントローラークラス
type CaribbeanStudCuiController struct {
	ci usecase.CaribbeanStudInteractorIF
}

// NewCaribbeanStudCuiController コンストラクタ
func NewCaribbeanStudCuiController(ci usecase.CaribbeanStudInteractorIF) *CaribbeanStudCuiController {
	return &CaribbeanStudCuiController{
		ci: ci,
	}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "b 100 10", "p", "f", "q"
func (cc *CaribbeanStudCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.ci.Reset() },
		[]string{"b", "bet", "p", "play", "f", "fold", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				ante, errMsg, ok := cuiutil.ParseIntArgKeys(args, "anteAmountRequired", "invalidAnteAmount", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				jackpot := 0
				if len(args) > 1 {
					jackpot, errMsg, ok = cuiutil.ParseIntArg(args[1:], "", "Invalid jackpot amount.", 0, math.MaxInt)
					if !ok {
						return errMsg, true
					}
				}
				return cc.ci.Bet(ante, jackpot), true
			case "p", "play":
				return cc.ci.Play(), true
			case "f", "fold":
				return cc.ci.Fold(), true
			case "h", "hint":
				return cc.ci.Hint(), true
			default:
				return handleCuiLog(cmd, cc.ci.ActionLog)
			}
		},
	)
}
