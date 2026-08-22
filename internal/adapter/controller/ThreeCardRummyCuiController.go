//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ThreeCardRummyCuiController スリーカード・ラミーCUIコントローラークラス
type ThreeCardRummyCuiController struct {
	ti usecase.ThreeCardRummyInteractorIF
}

// NewThreeCardRummyCuiController コンストラクタ
func NewThreeCardRummyCuiController(ti usecase.ThreeCardRummyInteractorIF) *ThreeCardRummyCuiController {
	return &ThreeCardRummyCuiController{
		ti: ti,
	}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "b 100 50", "p", "f", "q"
func (tcc *ThreeCardRummyCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return tcc.ti.Reset() },
		[]string{"b", "bet", "rb", "rebet", "p", "play", "f", "fold", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				ante, errMsg, ok := cuiutil.ParseIntArgKeys(args, "anteAmountRequired", "invalidAnteAmount", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				lowBonus := 0
				if len(args) > 1 {
					lowBonus, errMsg, ok = cuiutil.ParseIntArgKeys(args[1:], "", "invalidLowBonusAmount", 0, math.MaxInt)
					if !ok {
						return errMsg, true
					}
				}
				return tcc.ti.Bet(ante, lowBonus), true
			case "rb", "rebet":
				// **毎ラウンド同じ額を打ち直させない。** Web はワンクリックで
				// 再ベットできるのに、テキスト側は bet <ante> <lowBonus> を毎回
				// 手打ちする必要があった (#5513)。
				return tcc.ti.Rebet(), true
			case "p", "play":
				return tcc.ti.Play(), true
			case "f", "fold":
				return tcc.ti.Fold(), true
			default:
				return handleCuiHintAndLog(cmd, tcc.ti.Hint, tcc.ti.ActionLog)
			}
		},
	)
}
