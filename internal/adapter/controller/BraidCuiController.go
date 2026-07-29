//go:build !js || !wasm || extra2

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BraidCuiController ブレイド CUI コントローラークラス
type BraidCuiController struct {
	bi usecase.BraidInteractorIF
}

// NewBraidCuiController コンストラクタ
func NewBraidCuiController(bi usecase.BraidInteractorIF) *BraidCuiController {
	return &BraidCuiController{bi: bi}
}

// Exec コマンド実行
func (c *BraidCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.bi.Reset()
		},
		[]string{"d", "draw", "dir", "direction", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				return c.bi.Draw(), true
			case "dir", "direction":
				return c.handleDirection(args), true
			case "m", "move":
				return c.handleMove(args), true
			case "g", "giveup":
				return c.bi.GiveUp(), true
			case "ac", "autocomplete":
				return c.bi.AutoComplete(), true
			case "u", "undo":
				return c.bi.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.bi.Hint, c.bi.ActionLog)
			}
		},
	)
}

// handleDirection 積む向きを決める。`dir a` / `dir d`。
func (c *BraidCuiController) handleDirection(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("braid.promptDirection"), "dir {0}")
	}
	switch args[0] {
	case "a", "asc", "up":
		return c.bi.ChooseDirection(true)
	case "d", "desc", "down":
		return c.bi.ChooseDirection(false)
	default:
		return i18n.Tf("braid.invalidDirection", "val", args[0])
	}
}

// handleMove 移動コマンドを処理。supported syntax:
//
//	m br f              - the braid's tail to a foundation
//	m fd <idx> f        - a braid field to a foundation (refills from the braid)
//	m hp <idx> f        - a helper to a foundation
//	m w f               - the waste top to a foundation
//	m w hp <idx>        - the waste top into an empty helper
//
// 行き先が基礎札しかないので、`m ... t ...` のような枠間の移動は無い。
func (c *BraidCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("braid.promptSourceZone"), "m {0}")
	}
	switch args[0] {
	case "br":
		return c.handleMoveFromBraid(args[1:])
	case "fd":
		return c.handleMoveFromSlot(args[1:], "fd", c.bi.MoveFieldToFoundation)
	case "hp":
		return c.handleMoveFromSlot(args[1:], "hp", c.bi.MoveHelperToFoundation)
	case "w":
		return c.handleMoveFromWaste(args[1:])
	default:
		return i18n.Tf("braid.invalidFromZone", "val", args[0])
	}
}

// handleMoveFromBraid ブレイドの行き先は基礎札だけ。
func (c *BraidCuiController) handleMoveFromBraid(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("braid.promptBraidTo"), "m br {0}")
	}
	if args[0] != "f" {
		return i18n.Tf("braid.onlyToFoundation", "val", args[0])
	}
	return c.bi.MoveBraidToFoundation()
}

// handleMoveFromSlot ブレイド札・ヘルパーいずれも「枠番号 → 基礎札」で形が同じ。
func (c *BraidCuiController) handleMoveFromSlot(args []string, zone string, move func(int) string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("braid.promptSlotIdx"), "m "+zone+" {0}")
	}
	idx, err := strconv.Atoi(args[0])
	if err != nil {
		return i18n.Tf("braid.invalidSlot", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("braid.promptBraidTo"), "m "+zone+" "+args[0]+" {0}")
	}
	if args[1] != "f" {
		return i18n.Tf("braid.onlyToFoundation", "val", args[1])
	}
	return move(idx)
}

func (c *BraidCuiController) handleMoveFromWaste(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("braid.promptToZone"), "m w {0}")
	}
	switch args[0] {
	case "f":
		return c.bi.MoveWasteToFoundation()
	case "hp":
		if len(args) < 2 {
			return cuiutil.PromptRequest(i18n.T("braid.promptSlotIdx"), "m w hp {0}")
		}
		idx, err := strconv.Atoi(args[1])
		if err != nil {
			return i18n.Tf("braid.invalidSlot", "val", args[1])
		}
		return c.bi.MoveWasteToHelper(idx)
	default:
		return i18n.Tf("braid.invalidToZone", "val", args[0])
	}
}
