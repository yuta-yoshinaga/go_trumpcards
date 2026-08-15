//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SeahavenTowersCuiController シーヘイブンタワーズCUIコントローラークラス
type SeahavenTowersCuiController struct {
	si usecase.SeahavenTowersInteractorIF
}

// NewSeahavenTowersCuiController コンストラクタ
func NewSeahavenTowersCuiController(si usecase.SeahavenTowersInteractorIF) *SeahavenTowersCuiController {
	return &SeahavenTowersCuiController{si: si}
}

// Exec コマンド実行
func (c *SeahavenTowersCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(args []string) string {
			return c.si.Reset()
		},
		[]string{"m", "move", "f", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "f":
				return c.handleFoundationShorthand(args), true
			case "m", "move":
				return c.handleMove(args), true
			case "g", "giveup":
				return c.si.GiveUp(), true
			case "ac", "autocomplete":
				return c.si.AutoComplete(), true
			case "u", "undo":
				return c.si.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.si.Hint, c.si.ActionLog)
			}
		},
	)
}

// handleMove 移動コマンドを処理
func (c *SeahavenTowersCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("seahaventowers.promptSourceZone"), "m {0}")
	}
	// Shorthand: m <fromCol> [<toCol>] — tableau-to-tableau top card.
	if _, err := strconv.Atoi(args[0]); err == nil {
		return c.handleMoveShorthand(args)
	}
	from := args[0]
	if from != "t" && from != "c" {
		return invalidArg("seahaventowers.invalidFromZone", "val", from)
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

func (c *SeahavenTowersCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("seahaventowers.promptToZone"), fmt.Sprintf("m t %s {0}", args[0]))
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}

	switch args[1] {
	case "f":
		return c.si.MoveTableauToFoundation(fromCol)
	case "t":
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m t %d t {0}", fromCol))
		}
		toCol, err := strconv.Atoi(args[2])
		if err != nil {
			return invalidArg("invalidColumn", "val", args[2])
		}
		return c.si.MoveTableauToTableau(fromCol, -1, toCol)
	case "c":
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("promptCell"), fmt.Sprintf("m t %d c {0}", fromCol))
		}
		cell, err := strconv.Atoi(args[2])
		if err != nil {
			return invalidArg("invalidCell", "val", args[2])
		}
		return c.si.MoveTableauToFreeCell(fromCol, cell)
	default:
		cardIdx, err := strconv.Atoi(args[1])
		if err != nil {
			return i18n.T("seahaventowers.moveUsage")
		}
		if len(args) < 4 || args[2] != "t" {
			if len(args) == 3 && args[2] == "t" {
				return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m t %s %s t {0}", args[0], args[1]))
			}
			return i18n.T("seahaventowers.moveUsage")
		}
		toCol, err := strconv.Atoi(args[3])
		if err != nil {
			return invalidArg("invalidColumn", "val", args[3])
		}
		return c.si.MoveTableauToTableau(fromCol, cardIdx, toCol)
	}
}

func (c *SeahavenTowersCuiController) handleMoveFromFreeCell(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptCell"), "m c {0}")
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("seahaventowers.promptToZoneFromCell"), fmt.Sprintf("m c %s {0}", args[0]))
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
		return c.si.MoveFreeCellToTableau(cell, col)
	case "f":
		return c.si.MoveFreeCellToFoundation(cell)
	default:
		return invalidArg("seahaventowers.invalidToZone", "val", args[1])
	}
}

// handleFoundationShorthand handles `f <col>` (tableau-to-foundation).
func (c *SeahavenTowersCuiController) handleFoundationShorthand(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "f {0}")
	}
	col, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}
	return c.si.MoveTableauToFoundation(col)
}

func (c *SeahavenTowersCuiController) handleMoveShorthand(args []string) string {
	fromCol, _ := strconv.Atoi(args[0])
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m %s {0}", args[0]))
	}
	toCol, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[1])
	}
	return c.si.MoveTableauToTableau(fromCol, -1, toCol)
}
