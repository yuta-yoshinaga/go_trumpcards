//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MonteBankCuiController モンテバンクCUIコントローラークラス
type MonteBankCuiController struct {
	ci usecase.MonteBankInteractorIF
}

// NewMonteBankCuiController コンストラクタ
func NewMonteBankCuiController(ci usecase.MonteBankInteractorIF) *MonteBankCuiController {
	return &MonteBankCuiController{ci: ci}
}

// Exec ゲーム実行
// コマンド例: "r", "bet 1 50", "next", "hint", "log", "q"
//   - bet の第 1 引数は場札の番号 (1〜4)、第 2 引数は賭け金
func (cc *MonteBankCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.ci.Reset() },
		[]string{"bet", "b", "next", "hint", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				// **画面は 1 始まり、ドメインは 0 始まり。** 表示に合わせて
				// 受け取り、ここで 1 度だけ引く。
				idx, errMsg, ok := cuiutil.ParseIntArg(args,
					"Layout number is required.", "Invalid layout number. Please enter 1-4.",
					1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				bet, errMsg, ok := cuiutil.ParseIntArg(args[1:],
					"Bet is required.", "Invalid bet. Please enter a number.", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return cc.ci.PlaceBet(idx-1, bet), true
			case "next":
				return cc.ci.NextRound(), true
			case "hint":
				return cc.ci.Hint(), true
			default:
				return handleCuiLog(cmd, cc.ci.ActionLog)
			}
		},
	)
}
