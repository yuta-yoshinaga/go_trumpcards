//go:build !js || !wasm || extra4

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// AlaskaCuiController アラスカCUIコントローラークラス
type AlaskaCuiController struct {
	ri usecase.AlaskaInteractorIF
}

// NewAlaskaCuiController コンストラクタ
func NewAlaskaCuiController(ri usecase.AlaskaInteractorIF) *AlaskaCuiController {
	return &AlaskaCuiController{ri: ri}
}

// Exec コマンド実行
func (c *AlaskaCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.ri.Reset()
		},
		[]string{"m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "m", "move":
				return c.handleMove(args), true
			case "g", "giveup":
				return c.ri.GiveUp(), true
			case "ac", "autocomplete":
				return c.ri.AutoComplete(), true
			case "u", "undo":
				return c.ri.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.ri.Hint, c.ri.ActionLog)
			}
		},
	)
}

// handleMove 移動コマンドを処理
func (c *AlaskaCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("alaska.promptSourceColumn"), "m {0}")
	}

	// Shorthand: m <fromCol> [<toCol>] — tableau-to-tableau top card
	if _, err := strconv.Atoi(args[0]); err == nil {
		return c.handleMoveShorthand(args)
	}

	from := args[0]
	if from != "t" {
		return invalidArg("alaska.invalidFromZone", "val", from)
	}

	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}

	return c.handleMoveFromTableau(args[1:])
}

func (c *AlaskaCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("alaska.promptToZone"), fmt.Sprintf("m t %s {0}", args[0]))
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}

	if args[1] == "f" {
		return c.ri.MoveTableauToFoundation(fromCol)
	}

	// Format: m t <fromCol> <cardIdx> t <toCol>
	if len(args) < 4 || args[2] != "t" {
		if len(args) == 3 && args[2] == "t" {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m t %s %s t {0}", args[0], args[1]))
		}
		return i18n.MarkError(i18n.T("alaska.moveUsage"))
	}

	cardIdx, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidCardIndex", "val", args[1])
	}

	toCol, err := strconv.Atoi(args[3])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[3])
	}

	return c.ri.MoveTableauToTableau(fromCol, cardIdx, toCol)
}

func (c *AlaskaCuiController) handleMoveShorthand(args []string) string {
	fromCol, _ := strconv.Atoi(args[0])
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m %s {0}", args[0]))
	}
	toCol, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[1])
	}
	return c.ri.MoveTableauToTableau(fromCol, -1, toCol)
}
