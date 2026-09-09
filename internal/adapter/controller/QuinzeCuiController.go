//go:build !js || !wasm || extra2

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// QuinzeCuiController カーンズ CUI コントローラークラス
type QuinzeCuiController struct {
	si usecase.QuinzeInteractorIF
}

// NewQuinzeCuiController コンストラクタ
func NewQuinzeCuiController(si usecase.QuinzeInteractorIF) *QuinzeCuiController {
	return &QuinzeCuiController{si: si}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "h", "s", "h, s", "bh", "bs", "log", "q"
func (sc *QuinzeCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return sc.si.Reset() },
		[]string{"b", "bet", "deal", "h", "hit", "s", "stand", "bh", "bs", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args,
					"betAmountRequired", "invalidBetAmount",
					domain.QuinzeMinBet, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return sc.si.Bet(amount), true
			case "deal":
				return sc.si.Deal(), true
			case "h", "hit":
				return sc.si.Hit(), true
			case "s", "stand":
				return sc.si.Stand(), true
			case "bh":
				return sc.si.BankerHit(), true
			case "bs":
				return sc.si.BankerStand(), true
			default:
				if cmd == "log" || cmd == "l" {
					return sc.si.ActionLog(), true
				}
				return "", false
			}
		},
	)
}
