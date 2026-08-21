//go:build !js || !wasm || extra4

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// rankAndFileNoArgCommands maps no-arg CUI commands to RankAndFile methods.
var rankAndFileNoArgCommands = cuiutil.NewCommandMap[usecase.RankAndFileInteractorIF]().
	Add(usecase.RankAndFileInteractorIF.Draw, "d", "draw").
	Add(usecase.RankAndFileInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.RankAndFileInteractorIF.AutoComplete, "ac", "autocomplete").
	Add(usecase.RankAndFileInteractorIF.Undo, "u", "undo").
	Add(usecase.RankAndFileInteractorIF.Hint, "h", "hint").
	Add(usecase.RankAndFileInteractorIF.ActionLog, "log", "l")

// rankAndFileArgfulCommands lists alias names for argful commands handled in
// the Exec switch.
var rankAndFileArgfulCommands = []string{"m", "move"}

// RankAndFileCuiController ランク・アンド・ファイルCUIコントローラークラス
type RankAndFileCuiController struct {
	fi usecase.RankAndFileInteractorIF
}

// NewRankAndFileCuiController コンストラクタ
func NewRankAndFileCuiController(fi usecase.RankAndFileInteractorIF) *RankAndFileCuiController {
	return &RankAndFileCuiController{fi: fi}
}

// Exec コマンド実行
func (c *RankAndFileCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.fi.Reset()
		},
		append(rankAndFileNoArgCommands.Names(), rankAndFileArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := rankAndFileNoArgCommands.Lookup(cmd); ok {
				return fn(c.fi), true
			}
			switch cmd {
			case "m", "move":
				return c.handleMove(args), true
			default:
				return "", false
			}
		},
	)
}

// handleMove 移動コマンドを処理
func (c *RankAndFileCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("rankandfile.promptSourceZone"), "m {0}")
	}
	// Shorthand: m <fromCol> [<toCol>] — tableau-to-tableau top card
	if _, err := strconv.Atoi(args[0]); err == nil {
		return c.handleMoveShorthand(args)
	}
	from := args[0]
	if from != "w" && from != "t" {
		return invalidArg("rankandfile.invalidFromZone", "val", from)
	}
	if len(args) < 2 {
		switch from {
		case "w":
			return cuiutil.PromptRequest(i18n.T("rankandfile.promptToZone"), "m w {0}")
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

func (c *RankAndFileCuiController) handleMoveFromWaste(args []string) string {
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
		return c.fi.MoveWasteToTableau(col)
	case "f":
		return c.fi.MoveWasteToFoundation()
	default:
		return invalidArg("rankandfile.invalidToZone", "val", to)
	}
}

func (c *RankAndFileCuiController) handleMoveFromTableau(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("rankandfile.promptToZone"), fmt.Sprintf("m t %s {0}", args[0]))
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}

	if args[1] == "f" {
		return c.fi.MoveTableauToFoundation(fromCol)
	}

	// Format: m t <fromCol> <cardIdx> t <toCol>
	if len(args) < 4 || args[2] != "t" {
		if len(args) == 3 && args[2] == "t" {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m t %s %s t {0}", args[0], args[1]))
		}
		return i18n.MarkError(i18n.T("rankandfile.moveUsage"))
	}

	cardIdx, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidCardIndex", "val", args[1])
	}

	toCol, err := strconv.Atoi(args[3])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[3])
	}

	return c.fi.MoveTableauToTableau(fromCol, cardIdx, toCol)
}

func (c *RankAndFileCuiController) handleMoveShorthand(args []string) string {
	fromCol, _ := strconv.Atoi(args[0])
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m %s {0}", args[0]))
	}
	toCol, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[1])
	}
	return c.fi.MoveTableauToTableau(fromCol, -1, toCol)
}
