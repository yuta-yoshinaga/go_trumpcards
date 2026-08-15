//go:build !js || !wasm || solo

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// bristolNoArgCommands maps no-arg CUI commands to Bristol interactor methods.
var bristolNoArgCommands = cuiutil.NewCommandMap[usecase.BristolInteractorIF]().
	Add(usecase.BristolInteractorIF.Draw, "d", "draw").
	Add(usecase.BristolInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.BristolInteractorIF.AutoComplete, "ac", "autocomplete").
	Add(usecase.BristolInteractorIF.Undo, "u", "undo").
	Add(usecase.BristolInteractorIF.Hint, "h", "hint").
	Add(usecase.BristolInteractorIF.ActionLog, "log", "l")

// bristolArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var bristolArgfulCommands = []string{"m", "move"}

// BristolCuiController ブリストルCUIコントローラークラス
type BristolCuiController struct {
	bi usecase.BristolInteractorIF
}

// NewBristolCuiController コンストラクタ
func NewBristolCuiController(bi usecase.BristolInteractorIF) *BristolCuiController {
	return &BristolCuiController{bi: bi}
}

// Exec コマンド実行
func (c *BristolCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.bi.Reset() },
		append(bristolNoArgCommands.Names(), bristolArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := bristolNoArgCommands.Lookup(cmd); ok {
				return fn(c.bi), true
			}
			switch cmd {
			case "m", "move":
				return c.handleMove(args), true
			default:
				return "", false
			}
		},
	)
}

// handleMove 移動コマンドを処理する。
//
// 受理する書式:
//
//	m t <col> t <col>   タブロー列 → タブロー列
//	m t <col> f         タブロー列 → ファウンデーション
//	m n <fanIdx> t <col> ファン → タブロー列
//	m n <fanIdx> f      ファン → ファウンデーション
func (c *BristolCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("bristol.promptSourceZone"), "m {0}")
	}
	from := args[0]
	switch from {
	case "t":
		return c.handleMoveFromTableau(args[1:])
	case "n":
		return c.handleMoveFromFan(args[1:])
	default:
		return invalidArg("bristol.invalidFromZone", "val", from)
	}
}

func (c *BristolCuiController) handleMoveFromTableau(args []string) string {
	// args: ["<col>", "t", "<col>"] or ["<col>", "f"]
	if len(args) < 1 {
		return cuiutil.PromptRequest(i18n.T("bristol.promptFromColumn"), "m t {0}")
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("bristol.promptDestination"), "m t "+args[0]+" {0}")
	}
	switch args[1] {
	case "t":
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("bristol.promptToColumn"), "m t "+args[0]+" t {0}")
		}
		toCol, err := strconv.Atoi(args[2])
		if err != nil {
			return invalidArg("invalidColumn", "val", args[2])
		}
		return c.bi.MoveTableauToTableau(fromCol, toCol)
	case "f":
		return c.bi.MoveTableauToFoundation(fromCol)
	default:
		return i18n.MarkError(i18n.T("bristol.moveUsage"))
	}
}

func (c *BristolCuiController) handleMoveFromFan(args []string) string {
	// args: ["<fanIdx>", "t", "<col>"] or ["<fanIdx>", "f"]
	if len(args) < 1 {
		return cuiutil.PromptRequest(i18n.T("bristol.promptFan"), "m n {0}")
	}
	fanIdx, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("bristol.promptDestination"), "m n "+args[0]+" {0}")
	}
	switch args[1] {
	case "t":
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("bristol.promptToColumn"), "m n "+args[0]+" t {0}")
		}
		toCol, err := strconv.Atoi(args[2])
		if err != nil {
			return invalidArg("invalidColumn", "val", args[2])
		}
		return c.bi.MoveFanToTableau(fanIdx, toCol)
	case "f":
		return c.bi.MoveFanToFoundation(fanIdx)
	default:
		return i18n.MarkError(i18n.T("bristol.moveUsage"))
	}
}
