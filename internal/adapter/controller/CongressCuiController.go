//go:build !js || !wasm || extra3

package controller

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CongressCuiController コングレス CUI コントローラークラス
type CongressCuiController struct {
	ci usecase.CongressInteractorIF
}

// NewCongressCuiController コンストラクタ
func NewCongressCuiController(ci usecase.CongressInteractorIF) *CongressCuiController {
	return &CongressCuiController{ci: ci}
}

// Exec コマンド実行
func (c *CongressCuiController) Exec(command string) string {
	return execSolitaireCui(command, solitaireCuiFns{
		reset:        c.ci.Reset,
		move:         c.handleMove,
		giveUp:       c.ci.GiveUp,
		autoComplete: c.ci.AutoComplete,
		undo:         c.ci.Undo,
		hint:         c.ci.Hint,
		actionLog:    c.ci.ActionLog,
		extraCommands: map[string]func([]string) string{
			"d":    func([]string) string { return c.ci.Draw() },
			"draw": func([]string) string { return c.ci.Draw() },
		},
	})
}

// handleMove 移動コマンドを処理。supported syntax:
//
//	m t <pile> f        - a tableau top to a foundation
//	m t <from> t <to>   - one card between tableau piles
//	m w f               - the waste top to a foundation
//	m w t <pile>        - the waste top to a tableau pile
//	m s t <pile>        - the stock top straight into an empty pile
func (c *CongressCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("congress.promptSourceZone"), "m {0}")
	}
	switch args[0] {
	case "t":
		return c.handleMoveFromTableau(args[1:])
	case "w":
		return c.handleMoveFromWaste(args[1:])
	case "s":
		return c.handleMoveFromStock(args[1:])
	default:
		return invalidArg("congress.invalidFromZone", "val", args[0])
	}
}

func (c *CongressCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("congress.promptFromPile"), "m t {0}")
	}
	fromPile, msg, ok := cuiutil.ParseIntArgKeys(args, "", "congress.invalidPile", cuiutil.NoMin, cuiutil.NoMax)
	if !ok {
		return msg
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("congress.promptToZone"), fmt.Sprintf("m t %s {0}", args[0]))
	}
	switch args[1] {
	case "f":
		return c.ci.MoveTableauToFoundation(fromPile)
	case "t":
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("congress.promptToPile"), fmt.Sprintf("m t %s t {0}", args[0]))
		}
		toPile, msg, ok := cuiutil.ParseIntArgKeys(args[2:], "", "congress.invalidPile", cuiutil.NoMin, cuiutil.NoMax)
		if !ok {
			return msg
		}
		return c.ci.MoveTableauToTableau(fromPile, toPile)
	default:
		return invalidArg("congress.invalidToZone", "val", args[1])
	}
}

func (c *CongressCuiController) handleMoveFromWaste(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("congress.promptToZone"), "m w {0}")
	}
	switch args[0] {
	case "f":
		return c.ci.MoveWasteToFoundation()
	case "t":
		if len(args) < 2 {
			return cuiutil.PromptRequest(i18n.T("congress.promptToPile"), "m w t {0}")
		}
		pile, msg, ok := cuiutil.ParseIntArgKeys(args[1:], "", "congress.invalidPile", cuiutil.NoMin, cuiutil.NoMax)
		if !ok {
			return msg
		}
		return c.ci.MoveWasteToTableau(pile)
	default:
		return invalidArg("congress.invalidToZone", "val", args[0])
	}
}

// handleMoveFromStock 山札からは空き山を埋める手しかない。基礎札へ直接は送れない。
func (c *CongressCuiController) handleMoveFromStock(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("congress.promptToZone"), "m s {0}")
	}
	if args[0] != "t" {
		return invalidArg("congress.invalidToZone", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("congress.promptToPile"), "m s t {0}")
	}
	pile, msg, ok := cuiutil.ParseIntArgKeys(args[1:], "", "congress.invalidPile", cuiutil.NoMin, cuiutil.NoMax)
	if !ok {
		return msg
	}
	return c.ci.MoveStockToTableau(pile)
}
