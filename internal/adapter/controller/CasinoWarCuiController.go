//go:build !js || !wasm || extra4

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CasinoWarCuiController カジノウォーCUIコントローラークラス
type CasinoWarCuiController struct {
	ci usecase.CasinoWarInteractorIF
}

// NewCasinoWarCuiController コンストラクタ
func NewCasinoWarCuiController(ci usecase.CasinoWarInteractorIF) *CasinoWarCuiController {
	return &CasinoWarCuiController{ci: ci}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "surrender", "war", "log", "q"
func (cc *CasinoWarCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.ci.Reset() },
		[]string{"b", "bet", "surrender", "war", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args, "betAmountRequired", "invalidBetAmount", domain.CasinoWarMinBet, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return cc.ci.Bet(amount), true
			case "surrender":
				return cc.ci.Surrender(), true
			case "war":
				return cc.ci.War(), true
			default:
				return handleCuiLog(cmd, cc.ci.ActionLog)
			}
		},
	)
}
