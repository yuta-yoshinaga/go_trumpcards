//go:build !js || !wasm || extra2

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// AmericanToadCuiController アメリカン・トード CUI コントローラークラス
type AmericanToadCuiController struct {
	ai usecase.AmericanToadInteractorIF
}

// NewAmericanToadCuiController コンストラクタ
func NewAmericanToadCuiController(ai usecase.AmericanToadInteractorIF) *AmericanToadCuiController {
	return &AmericanToadCuiController{ai: ai}
}

// Exec コマンド実行
func (c *AmericanToadCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.ai.Reset()
		},
		[]string{"d", "draw", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				return c.ai.Draw(), true
			case "m", "move":
				return c.handleMove(args), true
			case "g", "giveup":
				return c.ai.GiveUp(), true
			case "ac", "autocomplete":
				return c.ai.AutoComplete(), true
			case "u", "undo":
				return c.ai.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.ai.Hint, c.ai.ActionLog)
			}
		},
	)
}

// handleMove 移動コマンドを処理。supported syntax:
//
//	m r f                     - reserve top to a foundation
//	m r t <col>               - reserve top to a tableau column
//	m w f                     - waste top to a foundation
//	m w t <col>               - waste top to a tableau column
//	m t <col> f               - tableau top to a foundation
//	m t <from> t <to> [<idx>] - tableau to tableau; <idx> is the head of the run
func (c *AmericanToadCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("americantoad.promptSourceZone"), "m {0}")
	}
	switch args[0] {
	case "r":
		return c.handleMoveFromPile(args[1:], "m r", c.ai.MoveReserveToFoundation, c.ai.MoveReserveToTableau)
	case "w":
		return c.handleMoveFromPile(args[1:], "m w", c.ai.MoveWasteToFoundation, c.ai.MoveWasteToTableau)
	case "t":
		return c.handleMoveFromTableau(args[1:])
	default:
		return i18n.Tf("americantoad.invalidFromZone", "val", args[0])
	}
}

// handleMoveFromPile リザーブと捨て札は「1 枚だけ使える山」で構文が同じなので、
// 送り先の 2 つの関数だけ差し替えて共有する。
func (c *AmericanToadCuiController) handleMoveFromPile(args []string, prefix string, toFoundation func() string, toTableau func(int) string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("americantoad.promptToZone"), prefix+" {0}")
	}
	switch args[0] {
	case "f":
		return toFoundation()
	case "t":
		if len(args) < 2 {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), prefix+" t {0}")
		}
		col, err := strconv.Atoi(args[1])
		if err != nil {
			return i18n.Tf("invalidColumn", "val", args[1])
		}
		return toTableau(col)
	default:
		return i18n.Tf("americantoad.invalidToZone", "val", args[0])
	}
}

func (c *AmericanToadCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("americantoad.promptToZone"), fmt.Sprintf("m t %s {0}", args[0]))
	}
	switch args[1] {
	case "f":
		return c.ai.MoveTableauToFoundation(fromCol)
	case "t":
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m t %s t {0}", args[0]))
		}
		toCol, err := strconv.Atoi(args[2])
		if err != nil {
			return i18n.Tf("invalidColumn", "val", args[2])
		}
		// 連番グループの先頭。省略時は -1 = 最上段 1 枚。
		cardIndex := -1
		if len(args) >= 4 {
			idx, err := strconv.Atoi(args[3])
			if err != nil {
				return i18n.Tf("americantoad.invalidCardIndex", "val", args[3])
			}
			cardIndex = idx
		}
		return c.ai.MoveTableauToTableau(fromCol, cardIndex, toCol)
	default:
		return i18n.Tf("americantoad.invalidToZone", "val", args[1])
	}
}
