//go:build !js || !wasm || extra

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KingoCuiController キンゴCUIコントローラークラス
type KingoCuiController struct {
	ci usecase.KingoInteractorIF
}

// NewKingoCuiController コンストラクタ
func NewKingoCuiController(ci usecase.KingoInteractorIF) *KingoCuiController {
	return &KingoCuiController{ci: ci}
}

// Exec ゲーム実行
// コマンド例: "r", "bet 20", "deal", "next", "hint", "log", "q"
func (cc *KingoCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.ci.Reset() },
		[]string{"bet", "b", "deal", "d", "next", "hint", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "bet", "b":
				// **額が要る手。** 省略は拒む。
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args,
					"amountRequired", "invalidAmountNotANumber", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return cc.ci.Bet(amount), true
			// **親の回は配るのが一手。** 張りとは別のコマンドにする。
			case "deal", "d":
				return cc.ci.Deal(), true
			case "next":
				return cc.ci.NextRound(), true
			case "hint":
				return cc.ci.Hint(), true
			default:
				return handleCuiLog(cmd, cc.ci.ActionLog)
			}
		},
	)
}
