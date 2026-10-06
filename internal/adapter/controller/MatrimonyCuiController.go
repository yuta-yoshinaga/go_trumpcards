//go:build !js || !wasm || extra

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MatrimonyCuiController マトリモニー CUI コントローラークラス
type MatrimonyCuiController struct {
	ci usecase.MatrimonyInteractorIF
}

// NewMatrimonyCuiController コンストラクタ
func NewMatrimonyCuiController(ci usecase.MatrimonyInteractorIF) *MatrimonyCuiController {
	return &MatrimonyCuiController{ci: ci}
}

// Exec コマンド実行
func (c *MatrimonyCuiController) Exec(command string) string {
	return execSolitaireCui(command, solitaireCuiFns{
		reset: c.ci.Reset, move: c.handleMove, giveUp: c.ci.GiveUp,
		autoComplete: c.ci.AutoComplete, undo: c.ci.Undo, hint: c.ci.Hint, actionLog: c.ci.ActionLog,
		extraCommands: map[string]func([]string) string{"d": func([]string) string { return c.ci.Draw() }, "draw": func([]string) string { return c.ci.Draw() }},
	})
}

// handleMove 移動コマンドを処理。supported syntax:
//
//	m t <slot>          - a tableau card to a foundation (its only move)
//	m w f               - the waste top to a foundation
//	m w t <slot>        - the waste top into an empty tableau slot
//	m s t <slot>        - the stock top straight into an empty tableau slot
func (c *MatrimonyCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("matrimony.promptSourceZone"), "m {0}")
	}
	switch args[0] {
	case "t":
		return c.handleMoveFromTableau(args[1:])
	case "w":
		return c.handleMoveFromWaste(args[1:])
	case "s":
		return c.handleMoveFromStock(args[1:])
	default:
		return invalidArg("matrimony.invalidFromZone", "val", args[0])
	}
}

// handleMoveFromTableau タブロー枠の札は基礎札へしか送れないので行き先を尋ねない。
// 「m t 3 f」と打たれても同じ手になるよう、末尾の f は受け流す。
func (c *MatrimonyCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("matrimony.promptFromPile"), "m t {0}")
	}
	slot, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("matrimony.invalidPile", "val", args[0])
	}
	if len(args) >= 2 && args[1] != "f" {
		return invalidArg("matrimony.invalidToZone", "val", args[1])
	}
	return c.ci.MoveTableauToFoundation(slot)
}

func (c *MatrimonyCuiController) handleMoveFromWaste(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("matrimony.promptToZone"), "m w {0}")
	}
	switch args[0] {
	case "f":
		return c.ci.MoveWasteToFoundation()
	case "t":
		if len(args) < 2 {
			return cuiutil.PromptRequest(i18n.T("matrimony.promptToPile"), "m w t {0}")
		}
		pile, err := strconv.Atoi(args[1])
		if err != nil {
			return invalidArg("matrimony.invalidPile", "val", args[1])
		}
		return c.ci.MoveWasteToTableau(pile)
	default:
		return invalidArg("matrimony.invalidToZone", "val", args[0])
	}
}

// handleMoveFromStock 山札からは空き山を埋める手しかない。基礎札へ直接は送れない。
func (c *MatrimonyCuiController) handleMoveFromStock(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("matrimony.promptToZone"), "m s {0}")
	}
	if args[0] != "t" {
		return invalidArg("matrimony.invalidToZone", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("matrimony.promptToPile"), "m s t {0}")
	}
	pile, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("matrimony.invalidPile", "val", args[1])
	}
	return c.ci.MoveStockToTableau(pile)
}
