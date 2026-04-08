package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FortyThievesCuiController フォーティシーブスCUIコントローラークラス
type FortyThievesCuiController struct {
	fi usecase.FortyThievesInteractorIF
}

// NewFortyThievesCuiController コンストラクタ
func NewFortyThievesCuiController(fi usecase.FortyThievesInteractorIF) *FortyThievesCuiController {
	return &FortyThievesCuiController{fi: fi}
}

// Exec コマンド実行
func (c *FortyThievesCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.fi.Reset()
		},
		[]string{"d", "draw", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				return c.fi.Draw(), true
			case "m", "move":
				return c.handleMove(args), true
			case "g", "giveup":
				return c.fi.GiveUp(), true
			case "ac", "autocomplete":
				return c.fi.AutoComplete(), true
			case "u", "undo":
				return c.fi.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.fi.Hint, c.fi.ActionLog)
			}
		},
	)
}

// handleMove 移動コマンドを処理
func (c *FortyThievesCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("fortythieves.promptSourceZone"), "m {0}")
	}
	from := args[0]
	if from != "w" && from != "t" {
		return i18n.Tf("fortythieves.invalidFromZone", "val", from)
	}
	if len(args) < 2 {
		switch from {
		case "w":
			return cuiutil.PromptRequest(i18n.T("fortythieves.promptToZone"), "m w {0}")
		case "t":
			return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
		}
	}
	switch from {
	case "w":
		return c.handleMoveFromWaste(args[1:])
	default: // "t"
		return c.handleMoveFromTableau(args[1:])
	}
}

func (c *FortyThievesCuiController) handleMoveFromWaste(args []string) string {
	to := args[0]
	switch to {
	case "t":
		if len(args) < 2 {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), "m w t {0}")
		}
		col, err := strconv.Atoi(args[1])
		if err != nil {
			return i18n.Tf("invalidColumn", "val", args[1])
		}
		return c.fi.MoveWasteToTableau(col)
	case "f":
		return c.fi.MoveWasteToFoundation()
	default:
		return i18n.Tf("fortythieves.invalidToZone", "val", to)
	}
}

func (c *FortyThievesCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("fortythieves.promptToZone"), fmt.Sprintf("m t %s {0}", args[0]))
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[0])
	}

	if args[1] == "f" {
		return c.fi.MoveTableauToFoundation(fromCol)
	}

	// Format: m t <fromCol> <cardIdx> t <toCol>
	if len(args) < 4 || args[2] != "t" {
		if len(args) == 3 && args[2] == "t" {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m t %s %s t {0}", args[0], args[1]))
		}
		return i18n.T("fortythieves.moveUsage")
	}

	cardIdx, err := strconv.Atoi(args[1])
	if err != nil {
		return i18n.Tf("invalidCardIndex", "val", args[1])
	}

	toCol, err := strconv.Atoi(args[3])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[3])
	}

	return c.fi.MoveTableauToTableau(fromCol, cardIdx, toCol)
}
