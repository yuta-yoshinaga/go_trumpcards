//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// IronCrossCuiController アイアンクロスCUIコントローラークラス
type IronCrossCuiController struct {
	ci usecase.IronCrossInteractorIF
}

// NewIronCrossCuiController コンストラクタ
func NewIronCrossCuiController(ci usecase.IronCrossInteractorIF) *IronCrossCuiController {
	return &IronCrossCuiController{ci: ci}
}

// Exec ゲーム実行
// コマンド例: "r", "check", "call", "bet 20", "raise 20", "fold",
// "vertical", "horizontal", "next", "hint", "log", "q"
func (cc *IronCrossCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.ci.Reset() },
		[]string{"fold", "f", "check", "k", "call", "c", "bet", "b", "raise",
			"vertical", "v", "horizontal", "h", "next", "hint", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "fold", "f":
				return cc.ci.Action(domain.IronCrossActionFold, 0), true
			case "check", "k":
				return cc.ci.Action(domain.IronCrossActionCheck, 0), true
			case "call", "c":
				return cc.ci.Action(domain.IronCrossActionCall, 0), true
			case "bet", "b", "raise":
				// **額が要る手。** 省略は拒む。
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args,
					"amountRequired", "invalidAmountNotANumber", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				action := domain.IronCrossActionBet
				if cmd == "raise" {
					action = domain.IronCrossActionRaise
				}
				return cc.ci.Action(action, amount), true
			case "vertical", "v":
				// **縦か横かは別コマンドにする。** ベットの手と同じ入口に混ぜると、
				// 一度きりの選択が打ち間違いで潰れる。
				return cc.ci.ChooseLine(int(domain.IronCrossLineVertical)), true
			case "horizontal", "h":
				return cc.ci.ChooseLine(int(domain.IronCrossLineHorizontal)), true
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
