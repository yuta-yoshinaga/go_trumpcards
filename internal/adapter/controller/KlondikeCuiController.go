package controller

import (
	"fmt"
	"strconv"

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
		func(_ []string) string {
			return c.ki.Reset()
		},
		unknownCommandMessage,
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				return c.ki.Draw(), true
			case "m", "move":
				return c.handleMove(args), true
			case "g", "giveup":
				return c.ki.GiveUp(), true
			case "h", "hint":
				return c.ki.Hint(), true
			case "ac", "autocomplete":
				return c.ki.AutoComplete(), true
			case "log", "l":
				return c.ki.ActionLog(), true
			}
			return "", false
		},
	)
}

// handleMove 移動コマンドを処理
func (c *KlondikeCuiController) handleMove(args []string) string {
	if len(args) < 2 {
		return "Usage: m w t <col> | m w f | m t <col> <idx> t <col> | m t <col> f"
	}
	from := args[0]
	switch from {
	case "w":
		return c.handleMoveFromWaste(args[1:])
	case "t":
		return c.handleMoveFromTableau(args[1:])
	default:
		return fmt.Sprintf("Invalid from zone: %s. Use 'w' (waste) or 't' (tableau).", from)
	}
}

func (c *KlondikeCuiController) handleMoveFromWaste(args []string) string {
	if len(args) < 1 {
		return "Usage: m w t <col> | m w f"
	}
	to := args[0]
	switch to {
	case "t":
		if len(args) < 2 {
			return "Usage: m w t <col>"
		}
		col, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Sprintf("Invalid column: %s.", args[1])
		}
		return c.ki.MoveWasteToTableau(col)
	case "f":
		return c.ki.MoveWasteToFoundation()
	default:
		return fmt.Sprintf("Invalid to zone: %s. Use 't' (tableau) or 'f' (foundation).", to)
	}
}

func (c *KlondikeCuiController) handleMoveFromTableau(args []string) string {
	if len(args) < 2 {
		return "Usage: m t <col> <idx> t <col> | m t <col> f"
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Sprintf("Invalid from column: %s.", args[0])
	}
	to := args[1]
	switch to {
	case "f":
		return c.ki.MoveTableauToFoundation(fromCol)
	case "t":
		return c.handleMoveTableauToTableau(fromCol, args[2:])
	default:
		// Check if it's a card index (e.g., "m t 0 3 t 4")
		cardIdx, err2 := strconv.Atoi(to)
		if err2 != nil {
			return fmt.Sprintf("Invalid to zone: %s.", to)
		}
		if len(args) < 4 {
			return "Usage: m t <col> <idx> t <col>"
		}
		if args[2] != "t" {
			return fmt.Sprintf("Invalid to zone: %s. Use 't' (tableau).", args[2])
		}
		toCol, err3 := strconv.Atoi(args[3])
		if err3 != nil {
			return fmt.Sprintf("Invalid to column: %s.", args[3])
		}
		return c.ki.MoveTableauToTableau(fromCol, cardIdx, toCol)
	}
}

func (c *KlondikeCuiController) handleMoveTableauToTableau(fromCol int, args []string) string {
	if len(args) < 1 {
		return "Usage: m t <col> t <col> (moves top card)"
	}
	toCol, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Sprintf("Invalid to column: %s.", args[0])
	}
	// Default to card index 0 (top face-up card) - simplified for "m t <from> t <to>"
	return c.ki.MoveTableauToTableau(fromCol, 0, toCol)
}
