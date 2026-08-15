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

// klondikeNoArgCommands maps no-arg CUI commands to Klondike interactor methods.
var klondikeNoArgCommands = cuiutil.NewCommandMap[usecase.KlondikeInteractorIF]().
	Add(usecase.KlondikeInteractorIF.Draw, "d", "draw").
	Add(usecase.KlondikeInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.KlondikeInteractorIF.AutoComplete, "ac", "autocomplete").
	Add(usecase.KlondikeInteractorIF.Undo, "u", "undo").
	Add(usecase.KlondikeInteractorIF.Hint, "h", "hint").
	Add(usecase.KlondikeInteractorIF.ActionLog, "log", "l")

// klondikeArgfulCommands lists alias names for argful commands handled in the
// Exec switch. "f" is included even though it can be no-arg, because it
// behaves differently with vs. without a column argument and isn't a clean
// nullary call.
var klondikeArgfulCommands = []string{"m", "move", "f"}

// KlondikeCuiController クロンダイクCUIコントローラークラス
type KlondikeCuiController struct {
	ki usecase.KlondikeInteractorIF
}

// NewKlondikeCuiController コンストラクタ
func NewKlondikeCuiController(ki usecase.KlondikeInteractorIF) *KlondikeCuiController {
	return &KlondikeCuiController{ki: ki}
}

// Exec コマンド実行
func (c *KlondikeCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(args []string) string {
			if len(args) > 0 {
				n, err := strconv.Atoi(args[0])
				if err == nil && (n == 1 || n == 3) {
					return c.ki.ResetWithConfig(domain.KlondikeConfig{DrawCount: n})
				}
			}
			return c.ki.Reset()
		},
		append(klondikeNoArgCommands.Names(), klondikeArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := klondikeNoArgCommands.Lookup(cmd); ok {
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
func (c *KlondikeCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("klondike.promptSourceZone"), "m {0}")
	}
	// Shorthand: m <fromCol> [<toCol>] — tableau-to-tableau top card
	if _, err := strconv.Atoi(args[0]); err == nil {
		return c.handleMoveShorthand(args)
	}
	from := args[0]
	if from != "w" && from != "t" {
		return invalidArg("klondike.invalidFromZone", "val", from)
	}
	if len(args) < 2 {
		switch from {
		case "w":
			return cuiutil.PromptRequest(i18n.T("klondike.promptToZone"), "m w {0}")
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

func (c *KlondikeCuiController) handleMoveFromWaste(args []string) string {
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
		return invalidArg("klondike.invalidToZone", "val", to)
	}
}

func (c *KlondikeCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("klondike.promptToZone"), fmt.Sprintf("m t %s {0}", args[0]))
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
		return i18n.T("klondike.moveUsage")
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
func (c *KlondikeCuiController) handleFoundationShorthand(args []string) string {
	if len(args) == 0 {
		return c.ki.MoveWasteToFoundation()
	}
	col, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}
	return c.ki.MoveTableauToFoundation(col)
}

func (c *KlondikeCuiController) handleMoveShorthand(args []string) string {
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
