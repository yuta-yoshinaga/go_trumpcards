//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DragonTigerCuiController ドラゴンタイガーCUIコントローラークラス
type DragonTigerCuiController struct {
	di usecase.DragonTigerInteractorIF
}

// NewDragonTigerCuiController コンストラクタ
func NewDragonTigerCuiController(di usecase.DragonTigerInteractorIF) *DragonTigerCuiController {
	return &DragonTigerCuiController{di: di}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100 d", "b 100 t", "b 100 e", "clear", "log", "q"
//   - dragon: d / dragon / 0
//   - tiger:  t / tiger  / 1
//   - tie:    e / tie    / 2
func (dc *DragonTigerCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return dc.di.Reset() },
		[]string{"b", "bet", "clear", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				if len(args) < 2 {
					return invalidArg("betAmountAndTypeRequired"), true
				}
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args[:1], "betAmountRequired", "invalidBetAmount", domain.DragonTigerMinBet, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				betType, ok := dragonTigerParseBetType(args[1])
				if !ok {
					return invalidArg("invalidBetTypeDragonTiger"), true
				}
				return dc.di.Bet(amount, betType), true
			case "clear":
				return dc.di.ClearHistory(), true
			default:
				return handleCuiLog(cmd, dc.di.ActionLog)
			}
		},
	)
}

// dragonTigerParseBetType ベットタイプを文字列から解析する。
func dragonTigerParseBetType(arg string) (int, bool) {
	switch arg {
	case "d", "dragon", "0":
		return domain.DragonTigerBetDragon, true
	case "t", "tiger", "1":
		return domain.DragonTigerBetTiger, true
	case "e", "tie", "2":
		return domain.DragonTigerBetTie, true
	default:
		return 0, false
	}
}
