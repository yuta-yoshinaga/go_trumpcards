//go:build !js || !wasm || extra3

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// StealingBundlesCuiController スティーリングバンドルCUIコントローラークラス
type StealingBundlesCuiController struct {
	si usecase.StealingBundlesInteractorIF
}

// NewStealingBundlesCuiController コンストラクタ
func NewStealingBundlesCuiController(si usecase.StealingBundlesInteractorIF) *StealingBundlesCuiController {
	return &StealingBundlesCuiController{si: si}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit          → ゲーム終了 ("bye.")
//	r / reset         → ゲームリセット (設定保持)
//	t / take <i>      → 手札の i 番目で場札を取る
//	s / steal <i> <p> → 手札の i 番目で席 p の束を奪う
//	d / trail <i>     → 手札の i 番目を場に置く (取れないときだけ)
//	g / giveup        → 投了
//	h / hint          → ヒント表示
//	log / l           → 棋譜表示
func (c *StealingBundlesCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.si.ResetWithConfig(c.si.GetConfig()) },
		[]string{"t", "take", "s", "steal", "d", "trail", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "t", "take":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
					cuiutil.NoMin, cuiutil.NoMax, c.si.Take)
			case "d", "trail":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
					cuiutil.NoMin, cuiutil.NoMax, c.si.Trail)
			case "s", "steal":
				return c.execSteal(args)
			case "g", "giveup":
				return c.si.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.si.Hint, c.si.ActionLog)
			}
		},
	)
}

// execSteal は steal コマンドの 2 引数を解釈する。
//
// **札と相手の両方が要ります。** どちらが欠けているかを言い分けます。
func (c *StealingBundlesCuiController) execSteal(args []string) (string, bool) {
	if len(args) < 1 {
		return invalidArg("cardIndexRequired"), true
	}
	cardIdx, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidCardIndex", "val", args[0]), true
	}
	if len(args) < 2 {
		return invalidArg("victimIndexRequired"), true
	}
	victim, err := strconv.Atoi(args[1])
	if err != nil {
		return "Invalid victim index: " + args[1] + ".", true
	}
	return c.si.Steal(cardIdx, victim), true
}
