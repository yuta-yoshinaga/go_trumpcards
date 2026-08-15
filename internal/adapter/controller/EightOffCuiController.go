//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// EightOffCuiController エイトオフCUIコントローラークラス
type EightOffCuiController struct {
	ei usecase.EightOffInteractorIF
}

// NewEightOffCuiController コンストラクタ
func NewEightOffCuiController(ei usecase.EightOffInteractorIF) *EightOffCuiController {
	return &EightOffCuiController{ei: ei}
}

// Exec コマンド実行
func (c *EightOffCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(args []string) string {
			return c.ei.Reset()
		},
		[]string{"m", "move", "f", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "f":
				return c.handleFoundationShorthand(args), true
			case "m", "move":
				return c.handleMove(args), true
			case "g", "giveup":
				return c.ei.GiveUp(), true
			case "ac", "autocomplete":
				return c.ei.AutoComplete(), true
			case "u", "undo":
				return c.ei.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.ei.Hint, c.ei.ActionLog)
			}
		},
	)
}

// handleMove 移動コマンドを処理
func (c *EightOffCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("eightoff.promptSourceZone"), "m {0}")
	}
	if _, err := strconv.Atoi(args[0]); err == nil {
		return c.handleMoveShorthand(args)
	}
	from := args[0]
	if from != "t" && from != "c" {
		return invalidArg("eightoff.invalidFromZone", "val", from)
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

func (c *EightOffCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("eightoff.promptToZone"), fmt.Sprintf("m t %s {0}", args[0]))
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}

	switch args[1] {
	case "f":
		return c.ei.MoveTableauToFoundation(fromCol)
	case "t":
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m t %d t {0}", fromCol))
		}
		toCol, err := strconv.Atoi(args[2])
		if err != nil {
			return invalidArg("invalidColumn", "val", args[2])
		}
		return c.ei.MoveTableauToTableau(fromCol, -1, toCol)
	case "c":
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("promptCell"), fmt.Sprintf("m t %d c {0}", fromCol))
		}
		cell, err := strconv.Atoi(args[2])
		if err != nil {
			return invalidArg("invalidCell", "val", args[2])
		}
		return c.ei.MoveTableauToFreeCell(fromCol, cell)
	default:
		cardIdx, err := strconv.Atoi(args[1])
		if err != nil {
			return i18n.T("eightoff.moveUsage")
		}
		if len(args) < 4 || args[2] != "t" {
			if len(args) == 3 && args[2] == "t" {
				return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m t %s %s t {0}", args[0], args[1]))
			}
			return i18n.T("eightoff.moveUsage")
		}
		toCol, err := strconv.Atoi(args[3])
		if err != nil {
			return invalidArg("invalidColumn", "val", args[3])
		}
		return c.ei.MoveTableauToTableau(fromCol, cardIdx, toCol)
	}
}

func (c *EightOffCuiController) handleMoveFromFreeCell(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptCell"), "m c {0}")
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("eightoff.promptToZoneFromCell"), fmt.Sprintf("m c %s {0}", args[0]))
	}
	cell, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidCell", "val", args[0])
	}

	switch args[1] {
	case "t":
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m c %d t {0}", cell))
		}
		col, err := strconv.Atoi(args[2])
		if err != nil {
			return invalidArg("invalidColumn", "val", args[2])
		}
		return c.ei.MoveFreeCellToTableau(cell, col)
	case "f":
		return c.ei.MoveFreeCellToFoundation(cell)
	default:
		return invalidArg("eightoff.invalidToZone", "val", args[1])
	}
}

// handleFoundationShorthand handles `f <col>` (tableau-to-foundation).
func (c *EightOffCuiController) handleFoundationShorthand(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "f {0}")
	}
	col, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}
	return c.ei.MoveTableauToFoundation(col)
}

func (c *EightOffCuiController) handleMoveShorthand(args []string) string {
	fromCol, _ := strconv.Atoi(args[0])
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m %s {0}", args[0]))
	}
	toCol, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[1])
	}
	return c.ei.MoveTableauToTableau(fromCol, -1, toCol)
}
