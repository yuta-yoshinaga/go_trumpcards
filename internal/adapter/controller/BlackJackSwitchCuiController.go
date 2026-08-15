//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BlackJackSwitchCuiController ブラックジャック・スイッチCUIコントローラー
type BlackJackSwitchCuiController struct {
	bi usecase.BlackJackSwitchInteractorIF
}

// NewBlackJackSwitchCuiController コンストラクタ
func NewBlackJackSwitchCuiController(bi usecase.BlackJackSwitchInteractorIF) *BlackJackSwitchCuiController {
	return &BlackJackSwitchCuiController{bi: bi}
}

// Exec ゲーム実行
//
// コマンド例:
//   - r / reset
//   - b 100 / bet 100
//   - sw / switch / k / keep
//   - h / hit / s / stand / dd / doubledown
//   - log / l
//   - q
func (bc *BlackJackSwitchCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return bc.bi.Reset() },
		[]string{"b", "bet", "sw", "switch", "k", "keep", "h", "hit", "s", "stand", "dd", "doubledown", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args, "betAmountRequired", "invalidBetAmount", domain.BJSwitchMinBet, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return bc.bi.Bet(amount), true
			case "sw", "switch":
				return bc.bi.Switch(), true
			case "k", "keep":
				return bc.bi.Keep(), true
			case "h", "hit":
				return bc.bi.Hit(), true
			case "s", "stand":
				return bc.bi.Stand(), true
			case "dd", "doubledown":
				return bc.bi.DoubleDown(), true
			default:
				return handleCuiLog(cmd, bc.bi.ActionLog)
			}
		},
	)
}
