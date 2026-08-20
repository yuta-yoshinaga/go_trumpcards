//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// AccordionCuiController アコーディオンCUIコントローラークラス
type AccordionCuiController struct {
	ai usecase.AccordionInteractorIF
}

// NewAccordionCuiController コンストラクタ
func NewAccordionCuiController(ai usecase.AccordionInteractorIF) *AccordionCuiController {
	return &AccordionCuiController{ai: ai}
}

// Exec コマンド実行
func (c *AccordionCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ai.Reset() },
		[]string{"m", "move", "g", "giveup", "h", "hint", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "m", "move":
				return c.handleMove(args), true
			case "g", "giveup":
				return c.ai.GiveUp(), true
			case "u", "undo":
				return c.ai.Undo(), true
			case "ac", "autocomplete":
				return c.ai.AutoComplete(), true
			default:
				return handleCuiHintAndLog(cmd, c.ai.Hint, c.ai.ActionLog)
			}
		},
	)
}

// handleMove 移動コマンドを処理
// 形式: m <fromIdx> <toIdx>
func (c *AccordionCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("accordion.promptFromIndex"), "m {0}")
	}
	fromIdx, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("accordion.promptToIndex"), fmt.Sprintf("m %d {0}", fromIdx))
	}
	toIdx, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[1])
	}
	return c.ai.Move(fromIdx, toIdx)
}
