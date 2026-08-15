//go:build !js || !wasm || extra2

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// AuldLangSyneCuiController オールド・ラング・サインCUIコントローラークラス
type AuldLangSyneCuiController struct {
	ci usecase.AuldLangSyneInteractorIF
}

// NewAuldLangSyneCuiController コンストラクタ
func NewAuldLangSyneCuiController(ci usecase.AuldLangSyneInteractorIF) *AuldLangSyneCuiController {
	return &AuldLangSyneCuiController{ci: ci}
}

// Exec コマンド実行
func (c *AuldLangSyneCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ci.Reset() },
		[]string{"d", "deal", "w", "g", "giveup", "h", "hint", "ac", "autocomplete", "u", "undo", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "deal":
				return c.ci.Deal(), true
			case "w":
				return c.handleWasteMove(args), true
			case "g", "giveup":
				return c.ci.GiveUp(), true
			case "ac", "autocomplete":
				return c.ci.AutoComplete(), true
			case "u", "undo":
				return c.ci.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.ci.Hint, c.ci.ActionLog)
			}
		},
	)
}

// handleWasteMove w <wasteIdx> f <fIdx>
func (c *AuldLangSyneCuiController) handleWasteMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("auldlangsyne.promptWasteIdx"), "w {0} f {1}")
	}
	wasteIdx, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[0])
	}
	if len(args) < 3 || args[1] != "f" {
		return cuiutil.PromptRequest(i18n.T("auldlangsyne.promptFoundationIdx"), fmt.Sprintf("w %d f {0}", wasteIdx))
	}
	fIdx, err := strconv.Atoi(args[2])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[2])
	}
	return c.ci.PlayWasteToFoundation(wasteIdx, fIdx)
}
