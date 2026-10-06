//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FreeCellCuiController フリーセルCUIコントローラークラス
type FreeCellCuiController struct {
	fi usecase.FreeCellInteractorIF
}

// NewFreeCellCuiController コンストラクタ
func NewFreeCellCuiController(fi usecase.FreeCellInteractorIF) *FreeCellCuiController {
	return &FreeCellCuiController{fi: fi}
}

// Exec コマンド実行
func (c *FreeCellCuiController) Exec(command string) string {
	return execSolitaireCui(command, solitaireCuiFns{
		reset:        c.fi.Reset,
		move:         c.handleMove,
		giveUp:       c.fi.GiveUp,
		autoComplete: c.fi.AutoComplete,
		undo:         c.fi.Undo,
		hint:         c.fi.Hint,
		actionLog:    c.fi.ActionLog,
		extraCommands: map[string]func([]string) string{
			"f": c.handleFoundationShorthand,
		},
	})
}

// handleMove 移動コマンドを処理
func (c *FreeCellCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("freecell.promptSourceZone"), "m {0}")
	}
	// Shorthand: m <fromCol> [<toCol>] — tableau-to-tableau top card
	if _, err := strconv.Atoi(args[0]); err == nil {
		return c.handleMoveShorthand(args)
	}
	from := args[0]
	if from != "t" && from != "c" {
		return invalidArg("freecell.invalidFromZone", "val", from)
	}
	if len(args) < 2 {
		switch from {
		case "t":
			return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
		case "c":
			return cuiutil.PromptRequest(i18n.T("promptCell"), "m c {0}")
		}
	}
	switch from {
	case "t":
		return c.handleMoveFromTableau(args[1:])
	default: // "c"
		return c.handleMoveFromFreeCell(args[1:])
	}
}

func (c *FreeCellCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("freecell.promptToZone"), fmt.Sprintf("m t %s {0}", args[0]))
	}
	fromCol, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args, "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
	if !parseOK {
		return parseMsg
	}

	switch args[1] {
	case "f":
		return c.fi.MoveTableauToFoundation(fromCol)
	case "t":
		// m t <fromCol> t <toCol> (top card move, cardIndex = last)
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m t %d t {0}", fromCol))
		}
		toCol, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[2:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
		if !parseOK {
			return parseMsg
		}
		// `m t <from> t <to>` names no card index, so it always means the top
		// card. The controller cannot resolve that to a real index -- it has no
		// board state -- so it passes -1 and the domain substitutes len-1.
		return c.fi.MoveTableauToTableau(fromCol, -1, toCol)
	case "c":
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("promptCell"), fmt.Sprintf("m t %d c {0}", fromCol))
		}
		cell, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[2:], "", "invalidCell", cuiutil.NoMin, cuiutil.NoMax)
		if !parseOK {
			return parseMsg
		}
		return c.fi.MoveTableauToFreeCell(fromCol, cell)
	default:
		// Could be: m t <fromCol> <cardIdx> t <toCol>
		cardIdx, err := strconv.Atoi(args[1])
		if err != nil {
			return i18n.MarkError(i18n.T("freecell.moveUsage"))
		}
		if len(args) < 4 || args[2] != "t" {
			if len(args) == 3 && args[2] == "t" {
				// Wizard state: m t <fromCol> <cardIdx> t — prompt for destination column
				return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m t %s %s t {0}", args[0], args[1]))
			}
			return i18n.MarkError(i18n.T("freecell.moveUsage"))
		}
		toCol, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[3:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
		if !parseOK {
			return parseMsg
		}
		return c.fi.MoveTableauToTableau(fromCol, cardIdx, toCol)
	}
}

func (c *FreeCellCuiController) handleMoveFromFreeCell(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptCell"), "m c {0}")
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("freecell.promptToZoneFromCell"), fmt.Sprintf("m c %s {0}", args[0]))
	}
	cell, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args, "", "invalidCell", cuiutil.NoMin, cuiutil.NoMax)
	if !parseOK {
		return parseMsg
	}

	switch args[1] {
	case "t":
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m c %d t {0}", cell))
		}
		col, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[2:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
		if !parseOK {
			return parseMsg
		}
		return c.fi.MoveFreeCellToTableau(cell, col)
	case "f":
		return c.fi.MoveFreeCellToFoundation(cell)
	default:
		return invalidArg("freecell.invalidToZone", "val", args[1])
	}
}

// handleFoundationShorthand handles `f <col>` (tableau-to-foundation).
func (c *FreeCellCuiController) handleFoundationShorthand(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "f {0}")
	}
	col, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args, "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
	if !parseOK {
		return parseMsg
	}
	return c.fi.MoveTableauToFoundation(col)
}

func (c *FreeCellCuiController) handleMoveShorthand(args []string) string {
	fromCol, _ := strconv.Atoi(args[0])
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m %s {0}", args[0]))
	}
	toCol, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[1:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
	if !parseOK {
		return parseMsg
	}
	return c.fi.MoveTableauToTableau(fromCol, -1, toCol)
}
