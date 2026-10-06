//go:build !js || !wasm || solo

package controller

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CrazyQuiltCuiController クレイジーキルト CUI コントローラークラス
type CrazyQuiltCuiController struct {
	ci usecase.CrazyQuiltInteractorIF
}

// NewCrazyQuiltCuiController コンストラクタ
func NewCrazyQuiltCuiController(ci usecase.CrazyQuiltInteractorIF) *CrazyQuiltCuiController {
	return &CrazyQuiltCuiController{ci: ci}
}

// Exec コマンド実行
func (c *CrazyQuiltCuiController) Exec(command string) string {
	return execSolitaireCui(command, solitaireCuiFns{
		reset:        c.ci.Reset,
		move:         c.handleMove,
		giveUp:       c.ci.GiveUp,
		autoComplete: c.ci.AutoComplete,
		undo:         c.ci.Undo,
		hint:         c.ci.Hint,
		actionLog:    c.ci.ActionLog,
		extraCommands: map[string]func([]string) string{
			"d":    func([]string) string { return c.ci.Draw() },
			"draw": func([]string) string { return c.ci.Draw() },
		},
	})
}

// handleMove 移動コマンドを処理。supported syntax:
//
//	m q <cell> f   - a quilt card to a foundation
//	m q <cell> w   - a quilt card onto the waste (one rank away, any suit)
//	m w f          - the waste top to a foundation
func (c *CrazyQuiltCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("crazyquilt.promptSourceZone"), "m {0}")
	}
	switch args[0] {
	case "q":
		return c.handleMoveFromQuilt(args[1:])
	case "w":
		return c.handleMoveFromWaste(args[1:])
	default:
		return invalidArg("crazyquilt.invalidFromZone", "val", args[0])
	}
}

// handleMoveFromQuilt キルトの札は基礎札か捨て札のどちらかへ送れる。
func (c *CrazyQuiltCuiController) handleMoveFromQuilt(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("crazyquilt.promptCell"), "m q {0}")
	}
	idx, msg, ok := cuiutil.ParseIntArgKeys(args, "", "crazyquilt.invalidCell", cuiutil.NoMin, cuiutil.NoMax)
	if !ok {
		return msg
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("crazyquilt.promptToZone"), fmt.Sprintf("m q %s {0}", args[0]))
	}
	switch args[1] {
	case "f":
		return c.ci.MoveQuiltToFoundation(idx)
	case "w":
		return c.ci.MoveQuiltToWaste(idx)
	default:
		return invalidArg("crazyquilt.invalidToZone", "val", args[1])
	}
}

// handleMoveFromWaste 捨て札は基礎札へしか送れないので行き先を尋ねない。
func (c *CrazyQuiltCuiController) handleMoveFromWaste(args []string) string {
	if len(args) >= 1 && args[0] != "f" {
		return invalidArg("crazyquilt.invalidToZone", "val", args[0])
	}
	return c.ci.MoveWasteToFoundation()
}
