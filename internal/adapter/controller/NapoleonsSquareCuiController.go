//go:build !js || !wasm || extra2

package controller

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NapoleonsSquareCuiController ナポレオンズ・スクエア CUI コントローラークラス
type NapoleonsSquareCuiController struct {
	ni usecase.NapoleonsSquareInteractorIF
}

// NewNapoleonsSquareCuiController コンストラクタ
func NewNapoleonsSquareCuiController(ni usecase.NapoleonsSquareInteractorIF) *NapoleonsSquareCuiController {
	return &NapoleonsSquareCuiController{ni: ni}
}

// Exec コマンド実行
func (c *NapoleonsSquareCuiController) Exec(command string) string {
	return execSolitaireCui(command, solitaireCuiFns{
		reset:        c.ni.Reset,
		move:         c.handleMove,
		giveUp:       c.ni.GiveUp,
		autoComplete: c.ni.AutoComplete,
		undo:         c.ni.Undo,
		hint:         c.ni.Hint,
		actionLog:    c.ni.ActionLog,
		extraCommands: map[string]func([]string) string{
			"d": func([]string) string { return c.ni.Draw() }, "draw": func([]string) string { return c.ni.Draw() },
		},
	})
}

// handleMove 移動コマンドを処理。supported syntax:
//
//	m w f                       - waste to a foundation
//	m w t <col>                 - waste to a tableau column
//	m t <col> f                 - tableau top to a foundation
//	m t <from> t <to> [<idx>]   - tableau to tableau; <idx> is the head of the
//	                              run to carry, defaulting to the top card
func (c *NapoleonsSquareCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("napoleonssquare.promptSourceZone"), "m {0}")
	}
	switch args[0] {
	case "w":
		return c.handleMoveFromWaste(args[1:])
	case "t":
		return c.handleMoveFromTableau(args[1:])
	default:
		return invalidArg("napoleonssquare.invalidFromZone", "val", args[0])
	}
}

func (c *NapoleonsSquareCuiController) handleMoveFromWaste(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("napoleonssquare.promptToZone"), "m w {0}")
	}
	switch args[0] {
	case "f":
		return c.ni.MoveWasteToFoundation()
	case "t":
		if len(args) < 2 {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), "m w t {0}")
		}
		col, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[1:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
		if !parseOK {
			return parseMsg
		}
		return c.ni.MoveWasteToTableau(col)
	default:
		return invalidArg("napoleonssquare.invalidToZone", "val", args[0])
	}
}

func (c *NapoleonsSquareCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	fromCol, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[0:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
	if !parseOK {
		return parseMsg
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("napoleonssquare.promptToZone"), fmt.Sprintf("m t %s {0}", args[0]))
	}
	switch args[1] {
	case "f":
		return c.ni.MoveTableauToFoundation(fromCol)
	case "t":
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m t %s t {0}", args[0]))
		}
		toCol, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[2:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
		if !parseOK {
			return parseMsg
		}
		// 連番グループの先頭。省略時は -1 = 最上段 1 枚。
		cardIndex := -1
		if len(args) >= 4 {
			idx, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[3:], "", "napoleonssquare.invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax)
			if !parseOK {
				return parseMsg
			}
			cardIndex = idx
		}
		return c.ni.MoveTableauToTableau(fromCol, cardIndex, toCol)
	default:
		return invalidArg("napoleonssquare.invalidToZone", "val", args[1])
	}
}
