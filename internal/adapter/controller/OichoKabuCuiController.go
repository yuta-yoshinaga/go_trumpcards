//go:build !js || !wasm || extra

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OichoKabuCuiController おいちょかぶCUIコントローラークラス
type OichoKabuCuiController struct {
	oi usecase.OichoKabuInteractorIF
}

// NewOichoKabuCuiController コンストラクタ
func NewOichoKabuCuiController(oi usecase.OichoKabuInteractorIF) *OichoKabuCuiController {
	return &OichoKabuCuiController{oi: oi}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "draw", "stand", "log", "q"
func (cc *OichoKabuCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.oi.Reset() },
		[]string{"b", "bet", "draw", "stand", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args, "betAmountRequired", "invalidBetAmount", domain.OichoKabuMinBet, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return cc.oi.Bet(amount), true
			case "d", "draw":
				return cc.oi.Draw(), true
			case "s", "stand":
				return cc.oi.Stand(), true
			default:
				return handleCuiLog(cmd, cc.oi.ActionLog)
			}
		},
	)
}
