//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// YukonCuiController ユーコンCUIコントローラークラス
type YukonCuiController struct {
	yi usecase.YukonInteractorIF
}

// NewYukonCuiController コンストラクタ
func NewYukonCuiController(yi usecase.YukonInteractorIF) *YukonCuiController {
	return &YukonCuiController{yi: yi}
}

// Exec コマンド実行
func (c *YukonCuiController) Exec(command string) string {
	return execSolitaireCui(command, solitaireCuiFns{
		reset: c.yi.Reset, move: c.handleMove, giveUp: c.yi.GiveUp,
		autoComplete: c.yi.AutoComplete, undo: c.yi.Undo, hint: c.yi.Hint, actionLog: c.yi.ActionLog,
	})
}

// handleMove 移動コマンドを処理
func (c *YukonCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("yukon.promptSourceColumn"), "m {0}")
	}

	// Shorthand: m <fromCol> [<toCol>] — tableau-to-tableau top card
	if _, err := strconv.Atoi(args[0]); err == nil {
		return c.handleMoveShorthand(args)
	}

	from := args[0]
	if from != "t" {
		return invalidArg("yukon.invalidFromZone", "val", from)
	}

	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}

	return c.handleMoveFromTableau(args[1:])
}

func (c *YukonCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("yukon.promptToZone"), fmt.Sprintf("m t %s {0}", args[0]))
	}
	fromCol, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[0:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
	if !parseOK {
		return parseMsg
	}

	if args[1] == "f" {
		return c.yi.MoveTableauToFoundation(fromCol)
	}

	// Format: m t <fromCol> <cardIdx> t <toCol>
	if len(args) < 4 || args[2] != "t" {
		if len(args) == 3 && args[2] == "t" {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m t %s %s t {0}", args[0], args[1]))
		}
		return i18n.MarkError(i18n.T("yukon.moveUsage"))
	}

	cardIdx, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[1:], "", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax)
	if !parseOK {
		return parseMsg
	}

	toCol, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[3:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
	if !parseOK {
		return parseMsg
	}

	return c.yi.MoveTableauToTableau(fromCol, cardIdx, toCol)
}

func (c *YukonCuiController) handleMoveShorthand(args []string) string {
	fromCol, _ := strconv.Atoi(args[0])
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m %s {0}", args[0]))
	}
	toCol, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[1:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
	if !parseOK {
		return parseMsg
	}
	return c.yi.MoveTableauToTableau(fromCol, -1, toCol)
}
