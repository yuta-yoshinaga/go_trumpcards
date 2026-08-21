//go:build !js || !wasm || extra2

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SlyFoxCuiController スライ・フォックス CUI コントローラークラス
type SlyFoxCuiController struct {
	ci usecase.SlyFoxInteractorIF
}

// NewSlyFoxCuiController コンストラクタ
func NewSlyFoxCuiController(ci usecase.SlyFoxInteractorIF) *SlyFoxCuiController {
	return &SlyFoxCuiController{ci: ci}
}

// Exec コマンド実行
func (c *SlyFoxCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.ci.Reset()
		},
		[]string{"d", "deal", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "deal":
				return c.handleDeal(args), true
			case "m", "move":
				return c.handleMove(args), true
			case "g", "giveup":
				return c.ci.GiveUp(), true
			case "ac", "autocomplete":
				return c.ci.AutoComplete(), true
			case "u", "undo":
				return c.ci.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.ci.Hint, c.ci.ActionLog)
			}
		},
	)
}

// handleDeal 配りコマンドを処理。supported syntax:
//
//	d <pile>            - deal the next card onto that reserve slot
//	d f <fIdx>          - deal it straight to a foundation (does not count)
//
// **捨て札は無いので、置き先を必ず言う。**引数なしの `d` は行き先を尋ねる。
func (c *SlyFoxCuiController) handleDeal(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("slyfox.promptDealTarget"), "d {0}")
	}
	if args[0] == "f" {
		if len(args) < 2 {
			return cuiutil.PromptRequest(i18n.T("slyfox.promptFoundationId"), "d f {0}")
		}
		fIdx, err := strconv.Atoi(args[1])
		if err != nil {
			return invalidArg("slyfox.invalidFoundation", "val", args[1])
		}
		return c.ci.DealToFoundation(fIdx)
	}
	pile, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("slyfox.invalidPile", "val", args[0])
	}
	return c.ci.DealToPile(pile)
}

// handleMove 移動コマンドを処理。supported syntax:
//
//	m t <pile>          - a reserve top to a foundation (its only legal move)
//
// **リザーブから組札へ送る手しか無い。**捨て札も、山札から空き枠を埋める手も
// このゲームには無いので、`m w ...` / `m s ...` は移動元として拒む。
func (c *SlyFoxCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("slyfox.promptSourceZone"), "m {0}")
	}
	if args[0] != "t" {
		return invalidArg("slyfox.invalidFromZone", "val", args[0])
	}
	return c.handleMoveFromTableau(args[1:])
}

// handleMoveFromTableau タブローの札は基礎札へしか送れないので、行き先を尋ねない。
// 「m t 3 f」と打たれても同じ手になるよう、末尾の f は受け流す。
func (c *SlyFoxCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("slyfox.promptFromPile"), "m t {0}")
	}
	fromPile, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("slyfox.invalidPile", "val", args[0])
	}
	if len(args) >= 2 && args[1] != "f" {
		return invalidArg("slyfox.invalidToZone", "val", args[1])
	}
	return c.ci.MoveTableauToFoundation(fromPile)
}
