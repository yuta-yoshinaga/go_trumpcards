//go:build !js || !wasm || classic

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ColoradoCuiController コロラド CUI コントローラークラス
type ColoradoCuiController struct {
	ci usecase.ColoradoInteractorIF
}

// NewColoradoCuiController コンストラクタ
func NewColoradoCuiController(ci usecase.ColoradoInteractorIF) *ColoradoCuiController {
	return &ColoradoCuiController{ci: ci}
}

// Exec コマンド実行
func (c *ColoradoCuiController) Exec(command string) string {
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
//	m t <pile>          - a tableau top to a foundation (its only legal move)
//	m w f               - the waste top to a foundation
//	m w t <pile>        - the waste top to a tableau pile
//	m s t <pile>        - the stock top straight into an empty pile
func (c *ColoradoCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("colorado.promptSourceZone"), "m {0}")
	}
	switch args[0] {
	case "t":
		return c.handleMoveFromTableau(args[1:])
	case "w":
		return c.handleMoveFromWaste(args[1:])
	case "s":
		return c.handleMoveFromStock(args[1:])
	default:
		return invalidArg("colorado.invalidFromZone", "val", args[0])
	}
}

// handleMoveFromTableau タブローの札は基礎札へしか送れないので、行き先を尋ねない。
// 「m t 3 f」と打たれても同じ手になるよう、末尾の f は受け流す。
func (c *ColoradoCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("colorado.promptFromPile"), "m t {0}")
	}
	fromPile, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("colorado.invalidPile", "val", args[0])
	}
	if len(args) >= 2 && args[1] != "f" {
		return invalidArg("colorado.invalidToZone", "val", args[1])
	}
	return c.ci.MoveTableauToFoundation(fromPile)
}

func (c *ColoradoCuiController) handleMoveFromWaste(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("colorado.promptToZone"), "m w {0}")
	}
	switch args[0] {
	case "f":
		return c.ci.MoveWasteToFoundation()
	case "t":
		if len(args) < 2 {
			return cuiutil.PromptRequest(i18n.T("colorado.promptToPile"), "m w t {0}")
		}
		pile, err := strconv.Atoi(args[1])
		if err != nil {
			return invalidArg("colorado.invalidPile", "val", args[1])
		}
		return c.ci.MoveWasteToTableau(pile)
	default:
		return invalidArg("colorado.invalidToZone", "val", args[0])
	}
}

// handleMoveFromStock 山札からは空き山を埋める手しかない。基礎札へ直接は送れない。
func (c *ColoradoCuiController) handleMoveFromStock(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("colorado.promptToZone"), "m s {0}")
	}
	if args[0] != "t" {
		return invalidArg("colorado.invalidToZone", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("colorado.promptToPile"), "m s t {0}")
	}
	pile, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("colorado.invalidPile", "val", args[1])
	}
	return c.ci.MoveStockToTableau(pile)
}
