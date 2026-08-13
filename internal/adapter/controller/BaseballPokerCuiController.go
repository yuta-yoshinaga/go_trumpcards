//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BaseballPokerCuiController ベースボールポーカーCUIコントローラークラス
type BaseballPokerCuiController struct {
	ci usecase.BaseballPokerInteractorIF
}

// NewBaseballPokerCuiController コンストラクタ
func NewBaseballPokerCuiController(ci usecase.BaseballPokerInteractorIF) *BaseballPokerCuiController {
	return &BaseballPokerCuiController{ci: ci}
}

// Exec ゲーム実行
// コマンド例: "r", "check", "call", "bet 20", "raise 20", "fold",
// "pay", "buyfold", "next", "hint", "log", "q"
func (cc *BaseballPokerCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.ci.Reset() },
		[]string{"fold", "f", "check", "k", "call", "c", "bet", "b", "raise",
			"pay", "p", "buyfold", "next", "hint", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "fold", "f":
				return cc.ci.Action(domain.BaseballActionFold, 0), true
			case "check", "k":
				return cc.ci.Action(domain.BaseballActionCheck, 0), true
			case "call", "c":
				return cc.ci.Action(domain.BaseballActionCall, 0), true
			case "bet", "b", "raise":
				// **額が要る手。** 省略は拒む。
				amount, errMsg, ok := cuiutil.ParseIntArg(args,
					"Amount is required.", "Invalid amount. Please enter a number.", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				action := domain.BaseballActionBet
				if cmd == "raise" {
					action = domain.BaseballActionRaise
				}
				return cc.ci.Action(action, amount), true
			// **買い増しの返事は名前で受ける。** 数値で受けると、送り忘れが
			// 「0 番の返事」= 支払いに化ける。
			case "pay", "p":
				return cc.ci.AnswerBuyIn(domain.BaseballBuyPay), true
			case "buyfold":
				return cc.ci.AnswerBuyIn(domain.BaseballBuyFold), true
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
