//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DoubleAttackBlackjackCuiController 追加ベット・ブラックジャックCUIコントローラークラス
type DoubleAttackBlackjackCuiController struct {
	ci usecase.DoubleAttackBlackjackInteractorIF
}

// NewDoubleAttackBlackjackCuiController コンストラクタ
func NewDoubleAttackBlackjackCuiController(ci usecase.DoubleAttackBlackjackInteractorIF) *DoubleAttackBlackjackCuiController {
	return &DoubleAttackBlackjackCuiController{ci: ci}
}

// Exec ゲーム実行
// コマンド例: "r", "bet 50 20", "attack 50", "hit", "stand", "double", "split", "next"
//   - bet の 2 つ目 (Bust It) は省略できる
//   - attack 0 は「見送り」
func (cc *DoubleAttackBlackjackCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.ci.Reset() },
		[]string{"bet", "b", "attack", "a", "hit", "h", "stand", "s", "double", "d",
			"split", "sp", "next", "hint", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				ante, errMsg, ok := cuiutil.ParseIntArg(args,
					"Ante is required.", "Invalid ante. Please enter a number.", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				bustIt := 0
				if len(args) > 1 {
					v, errMsg2, ok2 := cuiutil.ParseIntArg(args[1:],
						"", "Invalid bust it bet. Please enter a number.", 0, math.MaxInt)
					if !ok2 {
						return errMsg2, true
					}
					bustIt = v
				}
				return cc.ci.PlaceBet(ante, bustIt), true
			case "a", "attack":
				// **0 は「見送り」。** 上限はドメインが持つのでここでは書かない。
				amount, errMsg, ok := cuiutil.ParseIntArg(args,
					"Attack amount is required.", "Invalid amount. Please enter a number.", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return cc.ci.Attack(amount), true
			case "h", "hit":
				return cc.ci.Hit(), true
			case "s", "stand":
				return cc.ci.Stand(), true
			case "d", "double":
				return cc.ci.Double(), true
			case "sp", "split":
				return cc.ci.Split(), true
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
