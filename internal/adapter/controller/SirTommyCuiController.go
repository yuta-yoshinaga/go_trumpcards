//go:build !js || !wasm || extra2

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SirTommyCuiController サー・トミーCUIコントローラークラス
type SirTommyCuiController struct {
	ci usecase.SirTommyInteractorIF
}

// NewSirTommyCuiController コンストラクタ
func NewSirTommyCuiController(ci usecase.SirTommyInteractorIF) *SirTommyCuiController {
	return &SirTommyCuiController{ci: ci}
}

// Exec コマンド実行
func (c *SirTommyCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ci.Reset() },
		[]string{"s", "w", "g", "giveup", "h", "hint", "ac", "autocomplete", "u", "undo", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "s":
				return c.handleStockMove(args), true
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

// handleStockMove s <f|w> <idx>
func (c *SirTommyCuiController) handleStockMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("sirtommy.promptStockDestZone"), "s {0}")
	}
	dest := args[0]
	if dest != "f" && dest != "w" {
		return invalidArg("sirtommy.invalidDestZone", "val", dest)
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("sirtommy.promptIndex"), fmt.Sprintf("s %s {0}", dest))
	}
	idx, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[1])
	}
	if dest == "f" {
		return c.ci.PlayStockToFoundation(idx)
	}
	return c.ci.PlayStockToWaste(idx)
}

// handleWasteMove w <wasteIdx> f <fIdx>
func (c *SirTommyCuiController) handleWasteMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("sirtommy.promptWasteIdx"), "w {0} f {1}")
	}
	wasteIdx, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[0])
	}
	if len(args) < 3 || args[1] != "f" {
		return cuiutil.PromptRequest(i18n.T("sirtommy.promptFoundationIdx"), fmt.Sprintf("w %d f {0}", wasteIdx))
	}
	fIdx, err := strconv.Atoi(args[2])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[2])
	}
	return c.ci.PlayWasteToFoundation(wasteIdx, fIdx)
}
