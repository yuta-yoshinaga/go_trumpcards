package controller

import (
	"fmt"
	"strconv"

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
	return execCuiCommand(
		command,
		func(args []string) string {
			return c.fi.Reset()
		},
		[]string{"m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "m", "move":
				return c.handleMove(args), true
			case "g", "giveup":
				return c.fi.GiveUp(), true
			case "h", "hint":
				return c.fi.Hint(), true
			case "ac", "autocomplete":
				return c.fi.AutoComplete(), true
			case "log", "l":
				return c.fi.ActionLog(), true
			case "u", "undo":
				return c.fi.Undo(), true
			}
			return "", false
		},
	)
}

// handleMove 移動コマンドを処理
func (c *FreeCellCuiController) handleMove(args []string) string {
	if len(args) < 2 {
		return "Usage: m t <col> t <col> | m t <col> f | m t <col> c <cell> | m c <cell> t <col> | m c <cell> f"
	}
	from := args[0]
	switch from {
	case "t":
		return c.handleMoveFromTableau(args[1:])
	case "c":
		return c.handleMoveFromFreeCell(args[1:])
	default:
		return fmt.Sprintf("Invalid from zone: %s. Use 't' (tableau) or 'c' (freecell).", from)
	}
}

func (c *FreeCellCuiController) handleMoveFromTableau(args []string) string {
	if len(args) < 2 {
		return "Usage: m t <fromCol> f | m t <fromCol> <cardIdx> t <toCol> | m t <fromCol> c <cell>"
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Sprintf("Invalid from column: %s.", args[0])
	}

	switch args[1] {
	case "f":
		return c.fi.MoveTableauToFoundation(fromCol)
	case "t":
		// m t <fromCol> t <toCol> (top card move, cardIndex = last)
		if len(args) < 3 {
			return "Usage: m t <fromCol> t <toCol>"
		}
		toCol, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Sprintf("Invalid to column: %s.", args[2])
		}
		// Use cardIndex = -1 to signal top card; but the interactor expects actual index
		// For CUI simplicity, top card: cardIndex is len-1, but we don't know from here
		// Instead: use a large index that domain will reject, or pass last index
		// Actually, let's use the same pattern as Klondike: m t <col> <idx> t <col>
		// This branch handles: m t <fromCol> t <toCol> (move top card only)
		return c.fi.MoveTableauToTableau(fromCol, -1, toCol)
	case "c":
		if len(args) < 3 {
			return "Usage: m t <fromCol> c <cell>"
		}
		cell, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Sprintf("Invalid cell: %s.", args[2])
		}
		return c.fi.MoveTableauToFreeCell(fromCol, cell)
	default:
		// Could be: m t <fromCol> <cardIdx> t <toCol>
		cardIdx, err := strconv.Atoi(args[1])
		if err != nil {
			return "Invalid move command. Usage: m t <fromCol> f | m t <fromCol> <cardIdx> t <toCol> | m t <fromCol> c <cell>"
		}
		if len(args) < 4 || args[2] != "t" {
			return "Invalid move command. Usage: m t <fromCol> <cardIdx> t <toCol>"
		}
		toCol, err := strconv.Atoi(args[3])
		if err != nil {
			return fmt.Sprintf("Invalid to column: %s.", args[3])
		}
		return c.fi.MoveTableauToTableau(fromCol, cardIdx, toCol)
	}
}

func (c *FreeCellCuiController) handleMoveFromFreeCell(args []string) string {
	if len(args) < 2 {
		return "Usage: m c <cell> t <col> | m c <cell> f"
	}
	cell, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Sprintf("Invalid cell: %s.", args[0])
	}

	switch args[1] {
	case "t":
		if len(args) < 3 {
			return "Usage: m c <cell> t <col>"
		}
		col, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Sprintf("Invalid column: %s.", args[2])
		}
		return c.fi.MoveFreeCellToTableau(cell, col)
	case "f":
		return c.fi.MoveFreeCellToFoundation(cell)
	default:
		return fmt.Sprintf("Invalid to zone: %s. Use 't' (tableau) or 'f' (foundation).", args[1])
	}
}
