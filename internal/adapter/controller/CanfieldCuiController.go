package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CanfieldCuiController キャンフィールドCUIコントローラークラス
type CanfieldCuiController struct {
	ci usecase.CanfieldInteractorIF
}

// NewCanfieldCuiController コンストラクタ
func NewCanfieldCuiController(ci usecase.CanfieldInteractorIF) *CanfieldCuiController {
	return &CanfieldCuiController{ci: ci}
}

// Exec コマンド実行
func (c *CanfieldCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ci.Reset() },
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

// handleMove 移動コマンドを処理
func (c *CanfieldCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("canfield.promptSourceZone"), "m {0}")
	}
	from := args[0]
	if from != "w" && from != "t" && from != "r" {
		return i18n.Tf("canfield.invalidFromZone", "val", from)
	}
	if len(args) < 2 {
		switch from {
		case "w":
			return cuiutil.PromptRequest(i18n.T("canfield.promptToZone"), "m w {0}")
		case "r":
			return cuiutil.PromptRequest(i18n.T("canfield.promptToZone"), "m r {0}")
		case "t":
			return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
		}
	}
	switch from {
	case "w":
		return c.handleMoveFromWaste(args[1:])
	case "r":
		return c.handleMoveFromReserve(args[1:])
	default: // "t"
		return c.handleMoveFromTableau(args[1:])
	}
}

func (c *CanfieldCuiController) handleMoveFromWaste(args []string) string {
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
		return c.ci.MoveWasteToTableau(col)
	case "f":
		return c.ci.MoveWasteToFoundation()
	default:
		return i18n.Tf("canfield.invalidToZone", "val", to)
	}
}

func (c *CanfieldCuiController) handleMoveFromReserve(args []string) string {
	to := args[0]
	switch to {
	case "t":
		if len(args) < 2 {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), "m r t {0}")
		}
		col, err := strconv.Atoi(args[1])
		if err != nil {
			return i18n.Tf("invalidColumn", "val", args[1])
		}
		return c.ci.MoveReserveToTableau(col)
	case "f":
		return c.ci.MoveReserveToFoundation()
	default:
		return i18n.Tf("canfield.invalidToZone", "val", to)
	}
}

func (c *CanfieldCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("canfield.promptToZone"), fmt.Sprintf("m t %s {0}", args[0]))
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[0])
	}
	if args[1] == "f" {
		return c.ci.MoveTableauToFoundation(fromCol)
	}
	if len(args) < 4 || args[2] != "t" {
		if len(args) == 3 && args[2] == "t" {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m t %s %s t {0}", args[0], args[1]))
		}
		return i18n.T("canfield.moveUsage")
	}
	cardIdx, err := strconv.Atoi(args[1])
	if err != nil {
		return i18n.Tf("invalidCardIndex", "val", args[1])
	}
	toCol, err := strconv.Atoi(args[3])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[3])
	}
	return c.ci.MoveTableauToTableau(fromCol, cardIdx, toCol)
}
