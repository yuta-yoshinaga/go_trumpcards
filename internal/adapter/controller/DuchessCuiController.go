//go:build !js || !wasm || extra2

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DuchessCuiController ダッチェス CUI コントローラークラス
type DuchessCuiController struct {
	di usecase.DuchessInteractorIF
}

// NewDuchessCuiController コンストラクタ
func NewDuchessCuiController(di usecase.DuchessInteractorIF) *DuchessCuiController {
	return &DuchessCuiController{di: di}
}

// Exec コマンド実行
func (c *DuchessCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.di.Reset()
		},
		[]string{"b", "base", "d", "draw", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "base":
				return c.handleBase(args), true
			case "d", "draw":
				return c.di.Draw(), true
			case "m", "move":
				return c.handleMove(args), true
			case "g", "giveup":
				return c.di.GiveUp(), true
			case "ac", "autocomplete":
				return c.di.AutoComplete(), true
			case "u", "undo":
				return c.di.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}

// handleBase 開始ランクの選択: b <fan>
func (c *DuchessCuiController) handleBase(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("duchess.promptBaseFan"), "b {0}")
	}
	fan, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("duchess.invalidFanIdx", "val", args[0])
	}
	return c.di.ChooseBaseRank(fan)
}

// handleMove 移動コマンドを処理。supported syntax:
//
//	m r <fan> f               - reserve fan top to a foundation
//	m r <fan> t <col>         - reserve fan top to a tableau column
//	m w f                     - waste to a foundation
//	m w t <col>               - waste to a tableau column
//	m t <col> f               - tableau top to a foundation
//	m t <from> t <to> [<idx>] - tableau to tableau; <idx> is the head of the run
func (c *DuchessCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("duchess.promptSourceZone"), "m {0}")
	}
	switch args[0] {
	case "r":
		return c.handleMoveFromReserve(args[1:])
	case "w":
		return c.handleMoveFromWaste(args[1:])
	case "t":
		return c.handleMoveFromTableau(args[1:])
	default:
		return invalidArg("duchess.invalidFromZone", "val", args[0])
	}
}

func (c *DuchessCuiController) handleMoveFromReserve(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("duchess.promptBaseFan"), "m r {0}")
	}
	fan, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("duchess.invalidFanIdx", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("duchess.promptToZone"), fmt.Sprintf("m r %s {0}", args[0]))
	}
	switch args[1] {
	case "f":
		return c.di.MoveReserveToFoundation(fan)
	case "t":
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m r %s t {0}", args[0]))
		}
		col, err := strconv.Atoi(args[2])
		if err != nil {
			return invalidArg("invalidColumn", "val", args[2])
		}
		return c.di.MoveReserveToTableau(fan, col)
	default:
		return invalidArg("duchess.invalidToZone", "val", args[1])
	}
}

func (c *DuchessCuiController) handleMoveFromWaste(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("duchess.promptToZone"), "m w {0}")
	}
	switch args[0] {
	case "f":
		return c.di.MoveWasteToFoundation()
	case "t":
		if len(args) < 2 {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), "m w t {0}")
		}
		col, err := strconv.Atoi(args[1])
		if err != nil {
			return invalidArg("invalidColumn", "val", args[1])
		}
		return c.di.MoveWasteToTableau(col)
	default:
		return invalidArg("duchess.invalidToZone", "val", args[0])
	}
}

func (c *DuchessCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("duchess.promptToZone"), fmt.Sprintf("m t %s {0}", args[0]))
	}
	switch args[1] {
	case "f":
		return c.di.MoveTableauToFoundation(fromCol)
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
				return invalidArg("duchess.invalidCardIndex", "val", args[3])
			}
			cardIndex = idx
		}
		return c.di.MoveTableauToTableau(fromCol, cardIndex, toCol)
	default:
		return invalidArg("duchess.invalidToZone", "val", args[1])
	}
}
