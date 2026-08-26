//go:build !js || !wasm || extra

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SpeculationCuiController スペキュレーションCUIコントローラークラス
type SpeculationCuiController struct {
	ci usecase.SpeculationInteractorIF
}

// NewSpeculationCuiController コンストラクタ
func NewSpeculationCuiController(ci usecase.SpeculationInteractorIF) *SpeculationCuiController {
	return &SpeculationCuiController{ci: ci}
}

// Exec ゲーム実行
// コマンド例: "r", "f", "a", "d", "bid 50", "next", "hint", "log", "q"
//   - f はめくり、a/d は競りの受諾・拒否、bid は上乗せ額を指定して買う
func (cc *SpeculationCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.ci.Reset() },
		[]string{"f", "flip", "a", "accept", "d", "decline", "bid", "next", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "f", "flip":
				return cc.ci.Flip(), true
			case "a", "accept":
				return cc.ci.Accept(), true
			case "d", "decline":
				return cc.ci.Decline(), true
			case "bid":
				// **上乗せ額は必須。** 省略を 0 と読むと、断ったのか 0 で
				// 買おうとしたのか区別が付かない。
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args, "betRequired", "invalidBetANumber",
					1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return cc.ci.Bid(amount), true
			case "next":
				return cc.ci.NextRound(), true
			case "h", "hint":
				return cc.ci.Hint(), true
			default:
				return handleCuiLog(cmd, cc.ci.ActionLog)
			}
		},
	)
}
