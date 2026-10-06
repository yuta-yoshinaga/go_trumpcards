//go:build !js || !wasm || extra3

package controller

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SalicLawCuiController サリカ法典 CUI コントローラークラス
type SalicLawCuiController struct {
	ci usecase.SalicLawInteractorIF
}

// NewSalicLawCuiController コンストラクタ
func NewSalicLawCuiController(ci usecase.SalicLawInteractorIF) *SalicLawCuiController {
	return &SalicLawCuiController{ci: ci}
}

// Exec コマンド実行
func (c *SalicLawCuiController) Exec(command string) string {
	return execSolitaireCui(command, solitaireCuiFns{
		reset: c.ci.Reset, move: c.handleMove, giveUp: c.ci.GiveUp,
		autoComplete: c.ci.AutoComplete, undo: c.ci.Undo, hint: c.ci.Hint, actionLog: c.ci.ActionLog,
		extraCommands: map[string]func([]string) string{
			"d": func([]string) string { return c.ci.Draw() }, "draw": func([]string) string { return c.ci.Draw() },
		},
	})
}

// handleMove 移動コマンドを処理。supported syntax:
//
//	m t <pile> f        - a tableau top to a foundation
//	m t <from> t <to>   - one card between tableau piles
func (c *SalicLawCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("saliclaw.promptSourceZone"), "m {0}")
	}
	switch args[0] {
	case "t":
		return c.handleMoveFromTableau(args[1:])
	default:
		return invalidArg("saliclaw.invalidFromZone", "val", args[0])
	}
}

func (c *SalicLawCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("saliclaw.promptFromPile"), "m t {0}")
	}
	fromPile, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args, "", "saliclaw.invalidPile", cuiutil.NoMin, cuiutil.NoMax)
	if !parseOK {
		return parseMsg
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("saliclaw.promptToZone"), fmt.Sprintf("m t %s {0}", args[0]))
	}
	switch args[1] {
	case "f":
		return c.ci.MoveTableauToFoundation(fromPile)
	case "t":
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("saliclaw.promptToPile"), fmt.Sprintf("m t %s t {0}", args[0]))
		}
		toPile, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[2:], "", "saliclaw.invalidPile", cuiutil.NoMin, cuiutil.NoMax)
		if !parseOK {
			return parseMsg
		}
		return c.ci.MoveTableauToTableau(fromPile, toPile)
	default:
		return invalidArg("saliclaw.invalidToZone", "val", args[1])
	}
}
