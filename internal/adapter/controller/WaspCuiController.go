//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// waspNoArgCommands maps no-arg CUI commands to Wasp interactor methods.
var waspNoArgCommands = cuiutil.NewCommandMap[usecase.WaspInteractorIF]().
	Add(usecase.WaspInteractorIF.Deal, "d", "deal").
	Add(usecase.WaspInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.WaspInteractorIF.AutoComplete, "ac", "autocomplete").
	Add(usecase.WaspInteractorIF.Undo, "u", "undo").
	Add(usecase.WaspInteractorIF.Hint, "h", "hint").
	Add(usecase.WaspInteractorIF.ActionLog, "log", "l")

// waspArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var waspArgfulCommands = []string{"m", "move", "legal"}

// WaspCuiController ワスプCUIコントローラークラス
type WaspCuiController struct {
	si usecase.WaspInteractorIF
}

// NewWaspCuiController コンストラクタ
func NewWaspCuiController(si usecase.WaspInteractorIF) *WaspCuiController {
	return &WaspCuiController{si: si}
}

// Exec コマンド実行
func (c *WaspCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.si.Reset() },
		append(waspNoArgCommands.Names(), waspArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := waspNoArgCommands.Lookup(cmd); ok {
				return fn(c.si), true
			}
			switch cmd {
			case "m", "move":
				return c.handleMove(args), true
			case "legal":
				return c.handleLegal(args), true
			default:
				return "", false
			}
		},
	)
}

// handleMove 移動コマンドを処理
// Format: m t <fromCol> <cardIdx> t <toCol>
// Shorthand: m <fromCol> <toCol> (top card) or m <fromCol> <cardIdx> <toCol>
func (c *WaspCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("wasp.promptFromColumn"), "m {0}")
	}
	if _, err := strconv.Atoi(args[0]); err == nil {
		return c.handleMoveShorthand(args)
	}
	if args[0] != "t" {
		return i18n.T("wasp.moveUsage")
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("wasp.promptFromColumn"), "m t {0}")
	}
	fromCol, err := strconv.Atoi(args[1])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[1])
	}
	if len(args) < 3 {
		return cuiutil.PromptRequest(i18n.T("wasp.promptCardIndex"), fmt.Sprintf("m t %d {0} t", fromCol))
	}
	if len(args) < 5 || args[3] != "t" {
		if len(args) == 3 || (len(args) == 4 && args[3] == "t") {
			return cuiutil.PromptRequest(i18n.T("wasp.promptToColumn"), fmt.Sprintf("m t %d %s t {0}", fromCol, args[2]))
		}
		return i18n.T("wasp.moveUsage")
	}
	cardIdx, err := strconv.Atoi(args[2])
	if err != nil {
		return i18n.Tf("invalidCardIndex", "val", args[2])
	}
	toCol, err := strconv.Atoi(args[4])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[4])
	}
	return c.si.MoveTableauToTableau(fromCol, cardIdx, toCol)
}

// handleLegal previews the legal destination columns for the top card of the
// given column, including empty columns (which always accept a card in Wasp).
// Format: legal <col>
func (c *WaspCuiController) handleLegal(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("wasp.promptFromColumn"), "legal {0}")
	}
	col, err := strconv.Atoi(args[0])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[0])
	}
	return c.si.LegalMoves(col)
}

func (c *WaspCuiController) handleMoveShorthand(args []string) string {
	fromCol, _ := strconv.Atoi(args[0])
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("wasp.promptToColumn"), fmt.Sprintf("m %s {0}", args[0]))
	}
	if len(args) == 2 {
		toCol, err := strconv.Atoi(args[1])
		if err != nil {
			return i18n.Tf("invalidColumn", "val", args[1])
		}
		return c.si.MoveTableauToTableau(fromCol, -1, toCol)
	}
	cardIdx, err := strconv.Atoi(args[1])
	if err != nil {
		return i18n.Tf("invalidCardIndex", "val", args[1])
	}
	toCol, err := strconv.Atoi(args[2])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[2])
	}
	return c.si.MoveTableauToTableau(fromCol, cardIdx, toCol)
}
