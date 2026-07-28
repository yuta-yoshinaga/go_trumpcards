//go:build !js || !wasm || extra2

package controller

import (
	"fmt"
	"strconv"

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
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.ni.Reset()
		},
		[]string{"d", "draw", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				return c.ni.Draw(), true
			case "m", "move":
				return c.handleMove(args), true
			case "g", "giveup":
				return c.ni.GiveUp(), true
			case "ac", "autocomplete":
				return c.ni.AutoComplete(), true
			case "u", "undo":
				return c.ni.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.ni.Hint, c.ni.ActionLog)
			}
		},
	)
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
		return i18n.Tf("napoleonssquare.invalidFromZone", "val", args[0])
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
		col, err := strconv.Atoi(args[1])
		if err != nil {
			return i18n.Tf("invalidColumn", "val", args[1])
		}
		return c.ni.MoveWasteToTableau(col)
	default:
		return i18n.Tf("napoleonssquare.invalidToZone", "val", args[0])
	}
}

func (c *NapoleonsSquareCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[0])
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
		toCol, err := strconv.Atoi(args[2])
		if err != nil {
			return i18n.Tf("invalidColumn", "val", args[2])
		}
		// 連番グループの先頭。省略時は -1 = 最上段 1 枚。
		cardIndex := -1
		if len(args) >= 4 {
			idx, err := strconv.Atoi(args[3])
			if err != nil {
				return i18n.Tf("napoleonssquare.invalidCardIndex", "val", args[3])
			}
			cardIndex = idx
		}
		return c.ni.MoveTableauToTableau(fromCol, cardIndex, toCol)
	default:
		return i18n.Tf("napoleonssquare.invalidToZone", "val", args[1])
	}
}
