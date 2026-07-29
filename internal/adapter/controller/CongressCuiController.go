//go:build !js || !wasm || extra3

package controller

import (
	"fmt"
	"strconv"

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
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.ci.Reset()
		},
		[]string{"d", "draw", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				return c.ci.Draw(), true
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
		return i18n.Tf("congress.invalidFromZone", "val", args[0])
	}
}

func (c *CongressCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("congress.promptFromPile"), "m t {0}")
	}
	fromPile, err := strconv.Atoi(args[0])
	if err != nil {
		return i18n.Tf("congress.invalidPile", "val", args[0])
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
		toPile, err := strconv.Atoi(args[2])
		if err != nil {
			return i18n.Tf("congress.invalidPile", "val", args[2])
		}
		return c.ci.MoveTableauToTableau(fromPile, toPile)
	default:
		return i18n.Tf("congress.invalidToZone", "val", args[1])
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
		pile, err := strconv.Atoi(args[1])
		if err != nil {
			return i18n.Tf("congress.invalidPile", "val", args[1])
		}
		return c.ci.MoveWasteToTableau(pile)
	default:
		return i18n.Tf("congress.invalidToZone", "val", args[0])
	}
}

// handleMoveFromStock 山札からは空き山を埋める手しかない。基礎札へ直接は送れない。
func (c *CongressCuiController) handleMoveFromStock(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("congress.promptToZone"), "m s {0}")
	}
	if args[0] != "t" {
		return i18n.Tf("congress.invalidToZone", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("congress.promptToPile"), "m s t {0}")
	}
	pile, err := strconv.Atoi(args[1])
	if err != nil {
		return i18n.Tf("congress.invalidPile", "val", args[1])
	}
	return c.ci.MoveStockToTableau(pile)
}
