//go:build !js || !wasm || casino

package controller

import (
	"math"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ChinesePokerCuiController チャイニーズポーカーCUIコントローラークラス
type ChinesePokerCuiController struct {
	ci usecase.ChinesePokerInteractorIF
}

// NewChinesePokerCuiController コンストラクタ
func NewChinesePokerCuiController(ci usecase.ChinesePokerInteractorIF) *ChinesePokerCuiController {
	return &ChinesePokerCuiController{ci: ci}
}

// Exec ゲーム実行
func (cc *ChinesePokerCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.ci.Reset() },
		[]string{"b", "bet", "s", "set", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args, "betAmountRequired", "invalidBetAmount", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return cc.ci.Bet(amount), true
			case "s", "set":
				if len(args) < 8 {
					return "8 card indices are required (3 front + 5 middle).", true
				}
				frontIndices := make([]int, 3)
				for i := 0; i < 3; i++ {
					v, err := strconv.Atoi(args[i])
					if err != nil {
						return "Invalid front index: " + args[i], true
					}
					frontIndices[i] = v
				}
				middleIndices := make([]int, 5)
				for i := 0; i < 5; i++ {
					v, err := strconv.Atoi(args[3+i])
					if err != nil {
						return "Invalid middle index: " + args[3+i], true
					}
					middleIndices[i] = v
				}
				return cc.ci.SetHands(frontIndices, middleIndices), true
			case "h", "hint":
				return cc.ci.Hint(), true
			default:
				return handleCuiLog(cmd, cc.ci.ActionLog)
			}
		},
	)
}
