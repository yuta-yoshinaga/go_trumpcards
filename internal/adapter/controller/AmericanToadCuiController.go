//go:build !js || !wasm || extra2

package controller

import (
	"fmt"

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
	return execSolitaireCui(command, solitaireCuiFns{
		reset:        c.ai.Reset,
		move:         c.handleMove,
		giveUp:       c.ai.GiveUp,
		autoComplete: c.ai.AutoComplete,
		undo:         c.ai.Undo,
		hint:         c.ai.Hint,
		actionLog:    c.ai.ActionLog,
		extraCommands: map[string]func([]string) string{
			"d":    func([]string) string { return c.ai.Draw() },
			"draw": func([]string) string { return c.ai.Draw() },
		},
	})
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
		return invalidArg("americantoad.invalidFromZone", "val", args[0])
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
		col, msg, ok := cuiutil.ParseIntArgKeys(args[1:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
		if !ok {
			return msg
		}
		return toTableau(col)
	default:
		return invalidArg("americantoad.invalidToZone", "val", args[0])
	}
}

func (c *AmericanToadCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	fromCol, msg, ok := cuiutil.ParseIntArgKeys(args, "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
	if !ok {
		return msg
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
		toCol, msg, ok := cuiutil.ParseIntArgKeys(args[2:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
		if !ok {
			return msg
		}
		// 連番グループの先頭。省略時は -1 = 最上段 1 枚。
		cardIndex := -1
		if len(args) >= 4 {
			idx, msg, ok := cuiutil.ParseIntArgKeys(args[3:], "", "americantoad.invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax)
			if !ok {
				return msg
			}
			cardIndex = idx
		}
		return c.ai.MoveTableauToTableau(fromCol, cardIndex, toCol)
	default:
		return invalidArg("americantoad.invalidToZone", "val", args[1])
	}
}
