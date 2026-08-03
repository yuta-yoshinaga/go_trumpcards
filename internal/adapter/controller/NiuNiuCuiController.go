//go:build !js || !wasm || extra3

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NiuNiuCuiController 闘牛 CUI コントローラークラス
type NiuNiuCuiController struct {
	ni usecase.NiuNiuInteractorIF
}

// NewNiuNiuCuiController コンストラクタ
func NewNiuNiuCuiController(ni usecase.NiuNiuInteractorIF) *NiuNiuCuiController {
	return &NiuNiuCuiController{ni: ni}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "log", "q"
func (nc *NiuNiuCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return nc.ni.Reset() },
		[]string{"b", "bet", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				amount, errMsg, ok := cuiutil.ParseIntArg(args,
					"Bet amount is required.", "Invalid bet amount. Please enter a number.",
					domain.NiuNiuMinBet, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return nc.ni.Bet(amount), true
			default:
				if cmd == "log" || cmd == "l" {
					return nc.ni.ActionLog(), true
				}
				return "", false
			}
		},
	)
}
