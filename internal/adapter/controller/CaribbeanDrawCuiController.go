//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CaribbeanDrawCuiController カリビアン・ドロー・ポーカーCUIコントローラークラス
type CaribbeanDrawCuiController struct {
	ci usecase.CaribbeanDrawInteractorIF
}

// NewCaribbeanDrawCuiController コンストラクタ
func NewCaribbeanDrawCuiController(ci usecase.CaribbeanDrawInteractorIF) *CaribbeanDrawCuiController {
	return &CaribbeanDrawCuiController{
		ci: ci,
	}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "b 100 10", "d 1 3", "d", "p", "f", "q"
func (cc *CaribbeanDrawCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.ci.Reset() },
		[]string{"b", "bet", "d", "draw", "p", "play", "f", "fold", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				ante, errMsg, ok := cuiutil.ParseIntArgKeys(args, "anteAmountRequired", "invalidAnteAmount", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				jackpot := 0
				if len(args) > 1 {
					jackpot, errMsg, ok = cuiutil.ParseIntArgKeys(args[1:], "", "invalidJackpotAmount", 0, math.MaxInt)
					if !ok {
						return errMsg, true
					}
				}
				return cc.ci.Bet(ante, jackpot), true
			case "d", "draw":
				// **引数なしは「交換しない」。** 1 始まりの札番号で受け取る ——
				// 画面が 1 から数えて見せているのに 0 始まりで打たせると、
				// 意図と 1 枚ずれた札が消える。
				indices := make([]int, 0, len(args))
				for i := range args {
					n, errMsg, ok := cuiutil.ParseIntArgKeys(args[i:], "", "invalidCardIndex", 1, math.MaxInt)
					if !ok {
						return errMsg, true
					}
					indices = append(indices, n-1)
				}
				return cc.ci.Draw(indices), true
			case "p", "play":
				return cc.ci.Play(), true
			case "f", "fold":
				return cc.ci.Fold(), true
			case "h", "hint":
				return cc.ci.Hint(), true
			default:
				return handleCuiLog(cmd, cc.ci.ActionLog)
			}
		},
	)
}
