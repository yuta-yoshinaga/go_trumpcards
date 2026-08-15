//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// AndarBaharCuiController アンダーバハールCUIコントローラークラス
type AndarBaharCuiController struct {
	ai usecase.AndarBaharInteractorIF
}

// NewAndarBaharCuiController コンストラクタ
func NewAndarBaharCuiController(ai usecase.AndarBaharInteractorIF) *AndarBaharCuiController {
	return &AndarBaharCuiController{ai: ai}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100 a", "b 100 b 50 2", "clear", "hint", "log", "q"
//   - andar: a / andar / 0
//   - bahar: b / bahar / 1
//
// サイドベットは 3・4 番目の引数 (金額と帯 0-6) で、省略すればベットしません。
func (ac *AndarBaharCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return ac.ai.Reset() },
		[]string{"b", "bet", "clear", "hint", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				return ac.execBet(args)
			case "clear":
				return ac.ai.ClearHistory(), true
			case "hint":
				return ac.ai.Hint(), true
			default:
				return handleCuiLog(cmd, ac.ai.ActionLog)
			}
		},
	)
}

// execBet は "b <金額> <a|b> [サイド金額 サイド帯]" を解釈する。
func (ac *AndarBaharCuiController) execBet(args []string) (string, bool) {
	if len(args) < 2 {
		return invalidArg("betAmountAndTargetRequired"), true
	}
	amount, errMsg, ok := cuiutil.ParseIntArgKeys(args[:1],
		"betAmountRequired", "invalidBetAmount",
		domain.AndarBaharMinBet, math.MaxInt)
	if !ok {
		return errMsg, true
	}
	target, ok := andarBaharParseTarget(args[1])
	if !ok {
		return invalidArg("invalidBetTargetAndarBahar"), true
	}

	sideAmount, sideBand := 0, domain.AndarBaharSideNone
	if len(args) >= 4 {
		sideAmount, errMsg, ok = cuiutil.ParseIntArgKeys(args[2:3], "sideBetAmountRequired", "invalidSideBetAmountANumber",
			domain.AndarBaharMinBet, math.MaxInt)
		if !ok {
			return errMsg, true
		}
		sideBand, errMsg, ok = cuiutil.ParseIntArgKeys(args[3:4], "sideBetBandRequired", "invalidSideBetBandANumber",
			domain.AndarBaharSideFirst, domain.AndarBaharSide36Plus)
		if !ok {
			return errMsg, true
		}
	}
	return ac.ai.Bet(amount, target, sideAmount, sideBand), true
}

// andarBaharParseTarget はベット先を文字列から解析する。
func andarBaharParseTarget(arg string) (int, bool) {
	switch arg {
	case "a", "andar", "0":
		return domain.AndarBaharBetAndar, true
	case "b", "bahar", "1":
		return domain.AndarBaharBetBahar, true
	default:
		return 0, false
	}
}
