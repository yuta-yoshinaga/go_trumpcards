//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// scorpionNoArgCommands maps no-arg CUI commands to Scorpion interactor methods.
var scorpionNoArgCommands = cuiutil.NewCommandMap[usecase.ScorpionInteractorIF]().
	Add(usecase.ScorpionInteractorIF.Deal, "d", "deal").
	Add(usecase.ScorpionInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.ScorpionInteractorIF.AutoComplete, "ac", "autocomplete").
	Add(usecase.ScorpionInteractorIF.Hint, "h", "hint").
	Add(usecase.ScorpionInteractorIF.ActionLog, "log", "l")

// scorpionArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var scorpionArgfulCommands = []string{"m", "move", "legal", "u", "undo", "un", "undo_n"}

// ScorpionCuiController スコーピオンCUIコントローラークラス
type ScorpionCuiController struct {
	si usecase.ScorpionInteractorIF
}

// NewScorpionCuiController コンストラクタ
func NewScorpionCuiController(si usecase.ScorpionInteractorIF) *ScorpionCuiController {
	return &ScorpionCuiController{si: si}
}

// Exec コマンド実行
func (c *ScorpionCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.si.Reset() },
		append(scorpionNoArgCommands.Names(), scorpionArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := scorpionNoArgCommands.Lookup(cmd); ok {
				return fn(c.si), true
			}
			switch cmd {
			case "m", "move":
				return c.handleMove(args), true
			case "legal":
				return c.handleLegal(args), true
			case "u", "undo", "un", "undo_n":
				return c.handleUndo(cmd, args), true
			default:
				return "", false
			}
		},
	)
}

// handleUndo アンドゥコマンドを処理する。
// 引数なしの u / undo は 1 手戻す。
// 引数なしの un / undo_n は UndoToEscape() の手数を既定値として戻す。
// 引数ありは指定された手数を戻す。
func (c *ScorpionCuiController) handleUndo(cmd string, args []string) string {
	if len(args) == 0 {
		if cmd == "un" || cmd == "undo_n" {
			n := c.si.UndoToEscape()
			if n <= 0 {
				return i18n.MarkError(i18n.T("scorpion.noUndoToEscape"))
			}
			return c.si.UndoN(n)
		}
		return c.si.Undo()
	}
	out, _ := cuiutil.WithParsedIntKeys(args, "", "scorpion.invalidUndoCount", 1, cuiutil.NoMax, c.si.UndoN)
	return out
}

// handleMove 移動コマンドを処理
// Format: m t <fromCol> <cardIdx> t <toCol>
// Shorthand: m <fromCol> <toCol> (top card) or m <fromCol> <cardIdx> <toCol>
func (c *ScorpionCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("scorpion.promptFromColumn"), "m {0}")
	}
	if _, err := strconv.Atoi(args[0]); err == nil {
		return c.handleMoveShorthand(args)
	}
	if args[0] != "t" {
		return i18n.MarkError(i18n.T("scorpion.moveUsage"))
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("scorpion.promptFromColumn"), "m t {0}")
	}
	fromCol, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[1])
	}
	if len(args) < 3 {
		return cuiutil.PromptRequest(i18n.T("scorpion.promptCardIndex"), fmt.Sprintf("m t %d {0} t", fromCol))
	}
	if len(args) < 5 || args[3] != "t" {
		if len(args) == 3 || (len(args) == 4 && args[3] == "t") {
			return cuiutil.PromptRequest(i18n.T("scorpion.promptToColumn"), fmt.Sprintf("m t %d %s t {0}", fromCol, args[2]))
		}
		return i18n.MarkError(i18n.T("scorpion.moveUsage"))
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

// handleLegal previews the legal destination columns for the top card of the
// given column. In Scorpion an empty column only accepts a King.
// Format: legal <col>
func (c *ScorpionCuiController) handleLegal(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("scorpion.promptFromColumn"), "legal {0}")
	}
	col, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}
	return c.si.LegalMoves(col)
}

func (c *ScorpionCuiController) handleMoveShorthand(args []string) string {
	fromCol, _ := strconv.Atoi(args[0])
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("scorpion.promptToColumn"), fmt.Sprintf("m %s {0}", args[0]))
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
