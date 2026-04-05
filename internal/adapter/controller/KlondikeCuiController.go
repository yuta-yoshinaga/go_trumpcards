package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KlondikeCuiController クロンダイクCUIコントローラークラス
type KlondikeCuiController struct {
	ki usecase.KlondikeInteractorIF
}

// NewKlondikeCuiController コンストラクタ
func NewKlondikeCuiController(ki usecase.KlondikeInteractorIF) *KlondikeCuiController {
	return &KlondikeCuiController{ki: ki}
}

// Exec コマンド実行
func (c *KlondikeCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(args []string) string {
			if len(args) > 0 {
				n, err := strconv.Atoi(args[0])
				if err == nil && (n == 1 || n == 3) {
					return c.ki.ResetWithConfig(domain.KlondikeConfig{DrawCount: n})
				}
			}
			return c.ki.Reset()
		},
		[]string{"d", "draw", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				return c.ki.Draw(), true
			case "m", "move":
				return c.handleMove(args), true
			case "g", "giveup":
				return c.ki.GiveUp(), true
			case "ac", "autocomplete":
				return c.ki.AutoComplete(), true
			case "u", "undo":
				return c.ki.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.ki.Hint, c.ki.ActionLog)
			}
		},
	)
}

// handleMove 移動コマンドを処理
func (c *KlondikeCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("klondike.promptSourceZone"), "m {0}")
	}
	from := args[0]
	if from != "w" && from != "t" {
		return i18n.Tf("klondike.invalidFromZone", "val", from)
	}
	if len(args) < 2 {
		switch from {
		case "w":
			return cuiutil.PromptRequest(i18n.T("klondike.promptToZone"), "m w {0}")
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

func (c *KlondikeCuiController) handleMoveFromWaste(args []string) string {
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
		return c.ki.MoveWasteToTableau(col)
	case "f":
		return c.ki.MoveWasteToFoundation()
	default:
		return i18n.Tf("klondike.invalidToZone", "val", to)
	}
}

func (c *KlondikeCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("klondike.promptToZone"), fmt.Sprintf("m t %s {0}", args[0]))
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[0])
	}

	if args[1] == "f" {
		return c.ki.MoveTableauToFoundation(fromCol)
	}

	// The only other valid format is "m t <fromCol> <cardIdx> t <toCol>"
	if len(args) < 4 || args[2] != "t" {
		if len(args) == 3 && args[2] == "t" {
			// Wizard state: m t <fromCol> <cardIdx> t — prompt for destination column
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m t %s %s t {0}", args[0], args[1]))
		}
		return i18n.T("klondike.moveUsage")
	}

	cardIdx, err := strconv.Atoi(args[1])
	if err != nil {
		return i18n.Tf("invalidCardIndex", "val", args[1])
	}

	toCol, err := strconv.Atoi(args[3])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[3])
	}

	return c.ki.MoveTableauToTableau(fromCol, cardIdx, toCol)
}
