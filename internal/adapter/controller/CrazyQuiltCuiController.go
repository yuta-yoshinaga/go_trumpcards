//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

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
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.ci.Reset()
		},
		[]string{"d", "draw", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				return c.ci.Draw(), true
			case "m", "move":
				return c.handleMove(args), true
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
		return i18n.Tf("crazyquilt.invalidFromZone", "val", args[0])
	}
}

// handleMoveFromQuilt キルトの札は基礎札か捨て札のどちらかへ送れる。
func (c *CrazyQuiltCuiController) handleMoveFromQuilt(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("crazyquilt.promptCell"), "m q {0}")
	}
	idx, err := strconv.Atoi(args[0])
	if err != nil {
		return i18n.Tf("crazyquilt.invalidCell", "val", args[0])
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
		return i18n.Tf("crazyquilt.invalidToZone", "val", args[1])
	}
}

// handleMoveFromWaste 捨て札は基礎札へしか送れないので行き先を尋ねない。
func (c *CrazyQuiltCuiController) handleMoveFromWaste(args []string) string {
	if len(args) >= 1 && args[0] != "f" {
		return i18n.Tf("crazyquilt.invalidToZone", "val", args[0])
	}
	return c.ci.MoveWasteToFoundation()
}
