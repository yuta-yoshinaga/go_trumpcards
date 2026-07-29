//go:build !js || !wasm || extra2

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// WindmillCuiController ウィンドミル CUI コントローラークラス
type WindmillCuiController struct {
	wi usecase.WindmillInteractorIF
}

// NewWindmillCuiController コンストラクタ
func NewWindmillCuiController(wi usecase.WindmillInteractorIF) *WindmillCuiController {
	return &WindmillCuiController{wi: wi}
}

// Exec コマンド実行
func (c *WindmillCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.wi.Reset()
		},
		[]string{"d", "draw", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				return c.wi.Draw(), true
			case "m", "move":
				return c.handleMove(args), true
			case "g", "giveup":
				return c.wi.GiveUp(), true
			case "ac", "autocomplete":
				return c.wi.AutoComplete(), true
			case "u", "undo":
				return c.wi.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.wi.Hint, c.wi.ActionLog)
			}
		},
	)
}

// handleMove 移動コマンドを処理。supported syntax:
//
//	m s <sail> c        - a sail card to the centre foundation
//	m s <sail> k <idx>  - a sail card to a corner (King) foundation
//	m w c               - the waste top to the centre foundation
//	m w k <idx>         - the waste top to a corner foundation
//	m k <idx> c         - pull a corner top back onto the centre
func (c *WindmillCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("windmill.promptSourceZone"), "m {0}")
	}
	switch args[0] {
	case "s":
		return c.handleMoveFromSail(args[1:])
	case "w":
		return c.handleMoveFromWaste(args[1:])
	case "k":
		return c.handleMoveFromCorner(args[1:])
	default:
		return i18n.Tf("windmill.invalidFromZone", "val", args[0])
	}
}

func (c *WindmillCuiController) handleMoveFromSail(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("windmill.promptSail"), "m s {0}")
	}
	sail, err := strconv.Atoi(args[0])
	if err != nil {
		return i18n.Tf("windmill.invalidSailIdx", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("windmill.promptToZone"), fmt.Sprintf("m s %s {0}", args[0]))
	}
	switch args[1] {
	case "c":
		return c.wi.MoveSailToCenter(sail)
	case "k":
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("windmill.promptCorner"), fmt.Sprintf("m s %s k {0}", args[0]))
		}
		corner, err := strconv.Atoi(args[2])
		if err != nil {
			return i18n.Tf("windmill.invalidCornerIdx", "val", args[2])
		}
		return c.wi.MoveSailToCorner(sail, corner)
	default:
		return i18n.Tf("windmill.invalidToZone", "val", args[1])
	}
}

func (c *WindmillCuiController) handleMoveFromWaste(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("windmill.promptToZone"), "m w {0}")
	}
	switch args[0] {
	case "c":
		return c.wi.MoveWasteToCenter()
	case "k":
		if len(args) < 2 {
			return cuiutil.PromptRequest(i18n.T("windmill.promptCorner"), "m w k {0}")
		}
		corner, err := strconv.Atoi(args[1])
		if err != nil {
			return i18n.Tf("windmill.invalidCornerIdx", "val", args[1])
		}
		return c.wi.MoveWasteToCorner(corner)
	default:
		return i18n.Tf("windmill.invalidToZone", "val", args[0])
	}
}

// handleMoveFromCorner 四隅からの移動は中央への引き戻しだけ。降順の山へ戻す手は
// 規則上存在しないので、"m k <idx> k <idx>" は受け付けない。
func (c *WindmillCuiController) handleMoveFromCorner(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("windmill.promptCorner"), "m k {0}")
	}
	corner, err := strconv.Atoi(args[0])
	if err != nil {
		return i18n.Tf("windmill.invalidCornerIdx", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("windmill.promptToZone"), fmt.Sprintf("m k %s {0}", args[0]))
	}
	if args[1] != "c" {
		return i18n.Tf("windmill.invalidToZone", "val", args[1])
	}
	return c.wi.MoveCornerToCenter(corner)
}
