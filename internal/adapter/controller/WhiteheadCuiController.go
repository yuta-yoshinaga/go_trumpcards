//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// whiteheadNoArgCommands maps no-arg CUI commands to Whitehead interactor methods.
var whiteheadNoArgCommands = cuiutil.NewCommandMap[usecase.WhiteheadInteractorIF]().
	Add(usecase.WhiteheadInteractorIF.Draw, "d", "draw").
	Add(usecase.WhiteheadInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.WhiteheadInteractorIF.AutoComplete, "ac", "autocomplete").
	Add(usecase.WhiteheadInteractorIF.Undo, "u", "undo").
	Add(usecase.WhiteheadInteractorIF.Hint, "h", "hint").
	Add(usecase.WhiteheadInteractorIF.ActionLog, "log", "l")

// whiteheadArgfulCommands lists alias names for argful commands handled in the
// Exec switch. "f" is included even though it can be no-arg, because it
// behaves differently with vs. without a column argument and isn't a clean
// nullary call.
var whiteheadArgfulCommands = []string{"m", "move", "f"}

// WhiteheadCuiController ホワイトヘッドCUIコントローラークラス
type WhiteheadCuiController struct {
	ki usecase.WhiteheadInteractorIF
}

// NewWhiteheadCuiController コンストラクタ
func NewWhiteheadCuiController(ki usecase.WhiteheadInteractorIF) *WhiteheadCuiController {
	return &WhiteheadCuiController{ki: ki}
}

// Exec コマンド実行
func (c *WhiteheadCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(args []string) string {
			if len(args) > 0 {
				n, err := strconv.Atoi(args[0])
				if err == nil && (n == 1 || n == 3) {
					return c.ki.ResetWithConfig(domain.WhiteheadConfig{DrawCount: n})
				}
			}
			return c.ki.Reset()
		},
		append(whiteheadNoArgCommands.Names(), whiteheadArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := whiteheadNoArgCommands.Lookup(cmd); ok {
				return fn(c.ki), true
			}
			switch cmd {
			case "f":
				return c.handleFoundationShorthand(args), true
			case "m", "move":
				return c.handleMove(args), true
			default:
				return "", false
			}
		},
	)
}

// handleMove 移動コマンドを処理
func (c *WhiteheadCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("whitehead.promptSourceZone"), "m {0}")
	}
	// Shorthand: m <fromCol> [<toCol>] — tableau-to-tableau top card
	if _, err := strconv.Atoi(args[0]); err == nil {
		return c.handleMoveShorthand(args)
	}
	from := args[0]
	if from != "w" && from != "t" {
		return invalidArg("whitehead.invalidFromZone", "val", from)
	}
	if len(args) < 2 {
		switch from {
		case "w":
			return cuiutil.PromptRequest(i18n.T("whitehead.promptToZone"), "m w {0}")
		case "t":
			return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
		}
	}
	switch from {
	case "w":
		return c.handleMoveFromWaste(args[1:])
	default: // "t"
		return c.handleMoveFromTableau(args[1:])
	}
}

func (c *WhiteheadCuiController) handleMoveFromWaste(args []string) string {
	to := args[0]
	switch to {
	case "t":
		if len(args) < 2 {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), "m w t {0}")
		}
		col, err := strconv.Atoi(args[1])
		if err != nil {
			return invalidArg("invalidColumn", "val", args[1])
		}
		return c.ki.MoveWasteToTableau(col)
	case "f":
		return c.ki.MoveWasteToFoundation()
	default:
		return invalidArg("whitehead.invalidToZone", "val", to)
	}
}

func (c *WhiteheadCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("whitehead.promptToZone"), fmt.Sprintf("m t %s {0}", args[0]))
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}

	if args[1] == "f" {
		return c.ki.MoveTableauToFoundation(fromCol)
	}

	// The only other valid format is "m t <fromCol> <cardIdx> t <toCol>"
	if len(args) < 4 || args[2] != "t" {
		if len(args) == 3 && args[2] == "t" {
			// Wizard state: m t <fromCol> <cardIdx> t — prompt for destination column
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m t %s %s t {0}", args[0], args[1]))
		}
		return i18n.MarkError(i18n.T("whitehead.moveUsage"))
	}

	cardIdx, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidCardIndex", "val", args[1])
	}

	toCol, err := strconv.Atoi(args[3])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[3])
	}

	return c.ki.MoveTableauToTableau(fromCol, cardIdx, toCol)
}

// handleFoundationShorthand handles `f` (waste-to-foundation) and `f <col>` (tableau-to-foundation).
func (c *WhiteheadCuiController) handleFoundationShorthand(args []string) string {
	if len(args) == 0 {
		return c.ki.MoveWasteToFoundation()
	}
	col, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}
	return c.ki.MoveTableauToFoundation(col)
}

func (c *WhiteheadCuiController) handleMoveShorthand(args []string) string {
	fromCol, _ := strconv.Atoi(args[0])
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m %s {0}", args[0]))
	}
	toCol, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[1])
	}
	return c.ki.MoveTableauToTableau(fromCol, -1, toCol)
}
