//go:build !js || !wasm || extra2

package controller

import (
	"fmt"

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
	return execSolitaireCui(command, solitaireCuiFns{
		reset:        c.di.Reset,
		move:         c.handleMove,
		giveUp:       c.di.GiveUp,
		autoComplete: c.di.AutoComplete,
		undo:         c.di.Undo,
		hint:         c.di.Hint,
		actionLog:    c.di.ActionLog,
		extraCommands: map[string]func([]string) string{
			"b":    c.handleBase,
			"base": c.handleBase,
			"d":    func([]string) string { return c.di.Draw() },
			"draw": func([]string) string { return c.di.Draw() },
		},
	})
}

// handleBase 開始ランクの選択: b <fan>
func (c *DuchessCuiController) handleBase(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("duchess.promptBaseFan"), "b {0}")
	}
	fan, msg, ok := cuiutil.ParseIntArgKeys(args[0:], "", "duchess.invalidFanIdx", cuiutil.NoMin, cuiutil.NoMax)
	if !ok {
		return msg
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
	fan, msg, ok := cuiutil.ParseIntArgKeys(args[0:], "", "duchess.invalidFanIdx", cuiutil.NoMin, cuiutil.NoMax)
	if !ok {
		return msg
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
		col, msg, ok := cuiutil.ParseIntArgKeys(args[2:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
		if !ok {
			return msg
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
		col, msg, ok := cuiutil.ParseIntArgKeys(args[1:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
		if !ok {
			return msg
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
	fromCol, msg, ok := cuiutil.ParseIntArgKeys(args[0:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
	if !ok {
		return msg
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
		toCol, msg, ok := cuiutil.ParseIntArgKeys(args[2:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
		if !ok {
			return msg
		}
		// 連番グループの先頭。省略時は -1 = 最上段 1 枚。
		cardIndex := -1
		if len(args) >= 4 {
			idx, msg, ok := cuiutil.ParseIntArgKeys(args[3:], "", "duchess.invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax)
			if !ok {
				return msg
			}
			cardIndex = idx
		}
		return c.di.MoveTableauToTableau(fromCol, cardIndex, toCol)
	default:
		return invalidArg("duchess.invalidToZone", "val", args[1])
	}
}
