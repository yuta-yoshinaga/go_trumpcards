//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CincinnatiCuiController シンシナティCUIコントローラークラス
type CincinnatiCuiController struct {
	ci usecase.CincinnatiInteractorIF
}

// NewCincinnatiCuiController コンストラクタ
func NewCincinnatiCuiController(ci usecase.CincinnatiInteractorIF) *CincinnatiCuiController {
	return &CincinnatiCuiController{ci: ci}
}

// Exec ゲーム実行
// コマンド例: "r", "check", "call", "bet 20", "raise 20", "fold", "next", "hint", "log", "q"
func (cc *CincinnatiCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.ci.Reset() },
		[]string{"fold", "f", "check", "k", "call", "c", "bet", "b", "raise",
			"next", "hint", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "fold", "f":
				return cc.ci.Action(domain.CincinnatiActionFold, 0), true
			case "check", "k":
				return cc.ci.Action(domain.CincinnatiActionCheck, 0), true
			case "call", "c":
				return cc.ci.Action(domain.CincinnatiActionCall, 0), true
			case "bet", "b", "raise":
				// **額が要る手。** 省略は拒む。
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args,
					"amountRequired", "invalidAmountNotANumber", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				action := domain.CincinnatiActionBet
				if cmd == "raise" {
					action = domain.CincinnatiActionRaise
				}
				return cc.ci.Action(action, amount), true
			case "next":
				return cc.ci.NextHand(), true
			case "hint":
				return cc.ci.Hint(), true
			default:
				return handleCuiLog(cmd, cc.ci.ActionLog)
			}
		},
	)
}
