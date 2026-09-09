//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// willOTheWispNoArgCommands maps no-arg CUI commands to WillOTheWisp interactor methods.
var willOTheWispNoArgCommands = cuiutil.NewCommandMap[usecase.WillOTheWispInteractorIF]().
	Add(usecase.WillOTheWispInteractorIF.Deal, "d", "deal").
	Add(usecase.WillOTheWispInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.WillOTheWispInteractorIF.AutoComplete, "ac", "autocomplete").
	Add(usecase.WillOTheWispInteractorIF.Undo, "u", "undo").
	Add(usecase.WillOTheWispInteractorIF.Hint, "h", "hint").
	Add(usecase.WillOTheWispInteractorIF.ActionLog, "log", "l")

// willOTheWispArgfulCommands lists alias names for argful commands handled in the Exec switch.
var willOTheWispArgfulCommands = []string{"m", "move"}

// WillOTheWispCuiController ウィル・オ・ザ・ウィスプCUIコントローラークラス
type WillOTheWispCuiController struct {
	si usecase.WillOTheWispInteractorIF
}

// NewWillOTheWispCuiController コンストラクタ
func NewWillOTheWispCuiController(si usecase.WillOTheWispInteractorIF) *WillOTheWispCuiController {
	return &WillOTheWispCuiController{si: si}
}

// Exec コマンド実行
func (c *WillOTheWispCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.si.Reset() },
		append(willOTheWispNoArgCommands.Names(), willOTheWispArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := willOTheWispNoArgCommands.Lookup(cmd); ok {
				return fn(c.si), true
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

// handleMove 移動コマンドを処理
// Format: m t <fromCol> <cardIdx> t <toCol>
// Shorthand: m <fromCol> <toCol> (top card) or m <fromCol> <cardIdx> <toCol>
func (c *WillOTheWispCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m {0}")
	}
	if _, err := strconv.Atoi(args[0]); err == nil {
		return c.handleMoveShorthand(args)
	}
	if args[0] != "t" {
		return i18n.MarkError(i18n.T("willothewisp.moveUsage"))
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	fromCol, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[1])
	}
	if len(args) < 3 {
		return cuiutil.PromptRequest(i18n.T("promptCardIndex"), fmt.Sprintf("m t %d {0} t", fromCol))
	}
	if len(args) < 5 || args[3] != "t" {
		if len(args) == 3 || (len(args) == 4 && args[3] == "t") {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m t %d %s t {0}", fromCol, args[2]))
		}
		return i18n.MarkError(i18n.T("willothewisp.moveUsage"))
	}
	cardIdx, err := strconv.Atoi(args[2])
	if err != nil {
		return invalidArg("invalidCardIndex", "val", args[2])
	}
	toCol, err := strconv.Atoi(args[4])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[4])
	}
	return c.si.MoveTableauToTableau(fromCol, cardIdx, toCol)
}

func (c *WillOTheWispCuiController) handleMoveShorthand(args []string) string {
	fromCol, _ := strconv.Atoi(args[0])
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m %s {0}", args[0]))
	}
	if len(args) == 2 {
		toCol, err := strconv.Atoi(args[1])
		if err != nil {
			return invalidArg("invalidColumn", "val", args[1])
		}
		return c.si.MoveTableauToTableau(fromCol, -1, toCol)
	}
	cardIdx, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidCardIndex", "val", args[1])
	}
	toCol, err := strconv.Atoi(args[2])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[2])
	}
	return c.si.MoveTableauToTableau(fromCol, cardIdx, toCol)
}
