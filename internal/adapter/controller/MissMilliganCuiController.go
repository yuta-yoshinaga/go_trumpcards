//go:build !js || !wasm || extra2

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MissMilliganCuiController ミス・ミリガン CUI コントローラークラス
type MissMilliganCuiController struct {
	mi usecase.MissMilliganInteractorIF
}

// NewMissMilliganCuiController コンストラクタ
func NewMissMilliganCuiController(mi usecase.MissMilliganInteractorIF) *MissMilliganCuiController {
	return &MissMilliganCuiController{mi: mi}
}

// Exec コマンド実行
func (c *MissMilliganCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.mi.Reset()
		},
		[]string{"d", "deal", "m", "move", "wv", "waive", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "deal":
				return c.mi.Deal(), true
			case "m", "move":
				return c.handleMove(args), true
			case "wv", "waive":
				return c.handleWaive(args), true
			case "g", "giveup":
				return c.mi.GiveUp(), true
			case "ac", "autocomplete":
				return c.mi.AutoComplete(), true
			case "u", "undo":
				return c.mi.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.mi.Hint, c.mi.ActionLog)
			}
		},
	)
}

// handleWaive ウェイブコマンドを処理: wv <col> [<idx>]
func (c *MissMilliganCuiController) handleWaive(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "wv {0}")
	}
	col, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}
	cardIndex := -1
	if len(args) >= 2 {
		idx, err := strconv.Atoi(args[1])
		if err != nil {
			return invalidArg("missmilligan.invalidCardIndex", "val", args[1])
		}
		cardIndex = idx
	}
	return c.mi.Waive(col, cardIndex)
}

// handleMove 移動コマンドを処理。supported syntax:
//
//	m w f                     - waived card to a foundation
//	m w t <col>               - waived cards back onto a tableau column
//	m t <col> f               - tableau top to a foundation
//	m t <from> t <to> [<idx>] - tableau to tableau; <idx> is the head of the run
func (c *MissMilliganCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("missmilligan.promptSourceZone"), "m {0}")
	}
	switch args[0] {
	case "w":
		return c.handleMoveFromWaived(args[1:])
	case "t":
		return c.handleMoveFromTableau(args[1:])
	default:
		return invalidArg("missmilligan.invalidFromZone", "val", args[0])
	}
}

func (c *MissMilliganCuiController) handleMoveFromWaived(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("missmilligan.promptToZone"), "m w {0}")
	}
	switch args[0] {
	case "f":
		return c.mi.MoveWaivedToFoundation()
	case "t":
		if len(args) < 2 {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), "m w t {0}")
		}
		col, err := strconv.Atoi(args[1])
		if err != nil {
			return invalidArg("invalidColumn", "val", args[1])
		}
		return c.mi.PlaceWaived(col)
	default:
		return invalidArg("missmilligan.invalidToZone", "val", args[0])
	}
}

func (c *MissMilliganCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("missmilligan.promptToZone"), fmt.Sprintf("m t %s {0}", args[0]))
	}
	switch args[1] {
	case "f":
		return c.mi.MoveTableauToFoundation(fromCol)
	case "t":
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m t %s t {0}", args[0]))
		}
		toCol, err := strconv.Atoi(args[2])
		if err != nil {
			return invalidArg("invalidColumn", "val", args[2])
		}
		// 連番グループの先頭。省略時は -1 = 最上段 1 枚。
		cardIndex := -1
		if len(args) >= 4 {
			idx, err := strconv.Atoi(args[3])
			if err != nil {
				return invalidArg("missmilligan.invalidCardIndex", "val", args[3])
			}
			cardIndex = idx
		}
		return c.mi.MoveTableauToTableau(fromCol, cardIndex, toCol)
	default:
		return invalidArg("missmilligan.invalidToZone", "val", args[1])
	}
}
