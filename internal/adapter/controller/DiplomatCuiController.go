//go:build !js || !wasm || extra

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DiplomatCuiController ディプロマット CUI コントローラークラス
type DiplomatCuiController struct {
	ci usecase.DiplomatInteractorIF
}

// NewDiplomatCuiController コンストラクタ
func NewDiplomatCuiController(ci usecase.DiplomatInteractorIF) *DiplomatCuiController {
	return &DiplomatCuiController{ci: ci}
}

// Exec コマンド実行
func (c *DiplomatCuiController) Exec(command string) string {
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
//	m t <from> t <to>   - one card between tableau piles (an empty column too)
//	m w f               - the waste top to a foundation
//	m w t <pile>        - the waste top to a tableau pile (an empty column too)
func (c *DiplomatCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("diplomat.promptSourceZone"), "m {0}")
	}
	switch args[0] {
	case "t":
		return c.handleMoveFromTableau(args[1:])
	case "w":
		return c.handleMoveFromWaste(args[1:])
	default:
		return invalidArg("diplomat.invalidFromZone", "val", args[0])
	}
}

func (c *DiplomatCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("diplomat.promptFromPile"), "m t {0}")
	}
	fromPile, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("diplomat.invalidPile", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("diplomat.promptToZone"), fmt.Sprintf("m t %s {0}", args[0]))
	}
	switch args[1] {
	case "f":
		return c.ci.MoveTableauToFoundation(fromPile)
	case "t":
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("diplomat.promptToPile"), fmt.Sprintf("m t %s t {0}", args[0]))
		}
		toPile, err := strconv.Atoi(args[2])
		if err != nil {
			return invalidArg("diplomat.invalidPile", "val", args[2])
		}
		return c.ci.MoveTableauToTableau(fromPile, toPile)
	default:
		return invalidArg("diplomat.invalidToZone", "val", args[1])
	}
}

func (c *DiplomatCuiController) handleMoveFromWaste(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("diplomat.promptToZone"), "m w {0}")
	}
	switch args[0] {
	case "f":
		return c.ci.MoveWasteToFoundation()
	case "t":
		if len(args) < 2 {
			return cuiutil.PromptRequest(i18n.T("diplomat.promptToPile"), "m w t {0}")
		}
		pile, err := strconv.Atoi(args[1])
		if err != nil {
			return invalidArg("diplomat.invalidPile", "val", args[1])
		}
		return c.ci.MoveWasteToTableau(pile)
	default:
		return invalidArg("diplomat.invalidToZone", "val", args[0])
	}
}
