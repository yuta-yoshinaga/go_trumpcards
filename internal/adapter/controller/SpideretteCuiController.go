//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// spideretteNoArgCommands maps no-arg CUI commands to Spiderette interactor methods.
var spideretteNoArgCommands = cuiutil.NewCommandMap[usecase.SpideretteInteractorIF]().
	Add(usecase.SpideretteInteractorIF.Deal, "d", "deal").
	Add(usecase.SpideretteInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.SpideretteInteractorIF.AutoComplete, "ac", "autocomplete").
	Add(usecase.SpideretteInteractorIF.Undo, "u", "undo").
	Add(usecase.SpideretteInteractorIF.Hint, "h", "hint").
	Add(usecase.SpideretteInteractorIF.ActionLog, "log", "l")

// spideretteArgfulCommands lists alias names for argful commands handled in the Exec switch.
var spideretteArgfulCommands = []string{"m", "move"}

// SpideretteCuiController スパイダレットCUIコントローラークラス
type SpideretteCuiController struct {
	si usecase.SpideretteInteractorIF
}

// NewSpideretteCuiController コンストラクタ
func NewSpideretteCuiController(si usecase.SpideretteInteractorIF) *SpideretteCuiController {
	return &SpideretteCuiController{si: si}
}

// Exec コマンド実行
func (c *SpideretteCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.si.Reset() },
		append(spideretteNoArgCommands.Names(), spideretteArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := spideretteNoArgCommands.Lookup(cmd); ok {
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
func (c *SpideretteCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m {0}")
	}
	if _, err := strconv.Atoi(args[0]); err == nil {
		return c.handleMoveShorthand(args)
	}
	if args[0] != "t" {
		return i18n.T("spiderette.moveUsage")
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
		return i18n.T("spiderette.moveUsage")
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

func (c *SpideretteCuiController) handleMoveShorthand(args []string) string {
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
