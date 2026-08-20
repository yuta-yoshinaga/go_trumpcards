//go:build !js || !wasm || extra3

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TerraceCuiController テラス CUI コントローラークラス
type TerraceCuiController struct {
	ti usecase.TerraceInteractorIF
}

// NewTerraceCuiController コンストラクタ
func NewTerraceCuiController(ti usecase.TerraceInteractorIF) *TerraceCuiController {
	return &TerraceCuiController{ti: ti}
}

// Exec コマンド実行
func (c *TerraceCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.ti.Reset()
		},
		[]string{"d", "draw", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				return c.ti.Draw(), true
			case "m", "move":
				return c.handleMove(args), true
			case "g", "giveup":
				return c.ti.GiveUp(), true
			case "ac", "autocomplete":
				return c.ti.AutoComplete(), true
			case "u", "undo":
				return c.handleUndo(args), true
			default:
				return handleCuiHintAndLog(cmd, c.ti.Hint, c.ti.ActionLog)
			}
		},
	)
}

// handleUndo アンドゥ。引数なしなら 1 手、`undo <n>` なら n 手まとめて戻す。
//
// **手詰まりの案内が既に `undo <count>` と書いている** (terrace.undoToEscape)。
// 引数を捨てて 1 手だけ戻していたので、案内どおりに打った人は脱出したつもりで
// 手詰まりのまま置き去りにされていた (#5563)。
//
// 回数の上限は決めない。履歴より多ければドメインが答えるので、ここで打ち切ると
// Web の undo_n (素通し) と答えが食い違う。
func (c *TerraceCuiController) handleUndo(args []string) string {
	if len(args) == 0 {
		return c.ti.Undo()
	}
	out, _ := cuiutil.WithParsedIntKeys(args, "", "terrace.invalidUndoCount", 1, cuiutil.NoMax, c.ti.UndoN)
	return out
}

// handleMove 移動コマンドを処理。supported syntax:
//
//	m r f               - the terrace top to a foundation (its only destination)
//	m w f               - the waste top to a foundation
//	m w t <pile>        - the waste top to a tableau pile
//	m t <pile> f        - a tableau top to a foundation
//	m t <from> t <to>   - one card between tableau piles
func (c *TerraceCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("terrace.promptSourceZone"), "m {0}")
	}
	switch args[0] {
	case "r":
		return c.handleMoveFromReserve(args[1:])
	case "w":
		return c.handleMoveFromWaste(args[1:])
	case "t":
		return c.handleMoveFromTableau(args[1:])
	default:
		return invalidArg("terrace.invalidFromZone", "val", args[0])
	}
}

// handleMoveFromReserve テラスの行き先は基礎札だけ。タブローは受け付けない。
func (c *TerraceCuiController) handleMoveFromReserve(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("terrace.promptReserveTo"), "m r {0}")
	}
	if args[0] != "f" {
		return invalidArg("terrace.reserveOnlyToFoundation", "val", args[0])
	}
	return c.ti.MoveReserveToFoundation()
}

func (c *TerraceCuiController) handleMoveFromWaste(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("terrace.promptToZone"), "m w {0}")
	}
	switch args[0] {
	case "f":
		return c.ti.MoveWasteToFoundation()
	case "t":
		if len(args) < 2 {
			return cuiutil.PromptRequest(i18n.T("terrace.promptToPile"), "m w t {0}")
		}
		pile, err := strconv.Atoi(args[1])
		if err != nil {
			return invalidArg("terrace.invalidPile", "val", args[1])
		}
		return c.ti.MoveWasteToTableau(pile)
	default:
		return invalidArg("terrace.invalidToZone", "val", args[0])
	}
}

func (c *TerraceCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("terrace.promptFromPile"), "m t {0}")
	}
	fromPile, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("terrace.invalidPile", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("terrace.promptToZone"), fmt.Sprintf("m t %s {0}", args[0]))
	}
	switch args[1] {
	case "f":
		return c.ti.MoveTableauToFoundation(fromPile)
	case "t":
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("terrace.promptToPile"), fmt.Sprintf("m t %s t {0}", args[0]))
		}
		toPile, err := strconv.Atoi(args[2])
		if err != nil {
			return invalidArg("terrace.invalidPile", "val", args[2])
		}
		return c.ti.MoveTableauToTableau(fromPile, toPile)
	default:
		return invalidArg("terrace.invalidToZone", "val", args[1])
	}
}
