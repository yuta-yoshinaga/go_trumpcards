//go:build !js || !wasm || extra2

package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FaroCuiController はファロのCUIコントローラー。
type FaroCuiController struct {
	ci usecase.FaroInteractorIF
}

// NewFaroCuiController コンストラクタ。
func NewFaroCuiController(ci usecase.FaroInteractorIF) *FaroCuiController {
	return &FaroCuiController{ci: ci}
}

// Exec ゲーム実行。
// コマンド例: "r", "b 7 100", "b 7 100 c", "cb 7", "ca", "d", "call 3 9 12", "call", "n", "log", "q"
func (fc *FaroCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return fc.ci.Reset() },
		[]string{"b", "bet", "cb", "ca", "d", "deal", "call", "n", "next", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				return fc.handleBet(args), true
			case "cb":
				return fc.handleClearBet(args), true
			case "ca":
				return fc.ci.ClearAll(), true
			case "d", "deal":
				return fc.ci.DealTurn(), true
			case "call":
				return fc.handleCall(args), true
			case "n", "next":
				return fc.ci.NextRound(), true
			default:
				return handleCuiLog(cmd, fc.ci.ActionLog)
			}
		},
	)
}

// handleBet は "b <rank> <amount> [c]" を解析してベットを置く。
func (fc *FaroCuiController) handleBet(args []string) string {
	if len(args) < 2 {
		return i18n.MarkError(i18n.T("faro.errBetArgs"))
	}
	rank, err := strconv.Atoi(args[0])
	if err != nil {
		return i18n.MarkError(i18n.T("faro.errRank"))
	}
	amount, err := strconv.Atoi(args[1])
	if err != nil {
		return i18n.MarkError(i18n.T("faro.errAmount"))
	}
	copper := len(args) >= 3 && strings.EqualFold(args[2], "c")
	return fc.ci.PlaceBet(rank, amount, copper)
}

// handleClearBet は "cb <rank>" を解析してベットを取り消す。
func (fc *FaroCuiController) handleClearBet(args []string) string {
	if len(args) < 1 {
		return i18n.MarkError(i18n.T("faro.errRank"))
	}
	rank, err := strconv.Atoi(args[0])
	if err != nil {
		return i18n.MarkError(i18n.T("faro.errRank"))
	}
	return fc.ci.ClearBet(rank)
}

// handleCall は "call [<r1> <r2> <r3>]" を解析してコールする（引数なしで見送り）。
func (fc *FaroCuiController) handleCall(args []string) string {
	if len(args) == 0 {
		return fc.ci.Call(nil)
	}
	order := make([]int, 0, len(args))
	for _, a := range args {
		r, err := strconv.Atoi(a)
		if err != nil {
			return i18n.MarkError(i18n.T("faro.errCallArgs"))
		}
		order = append(order, r)
	}
	return fc.ci.Call(order)
}
