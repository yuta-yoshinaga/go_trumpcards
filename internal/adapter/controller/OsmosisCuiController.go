//go:build !js || !wasm || solo

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// osmosisNoArgCommands maps no-arg CUI commands to Osmosis interactor methods.
var osmosisNoArgCommands = cuiutil.NewCommandMap[usecase.OsmosisInteractorIF]().
	Add(usecase.OsmosisInteractorIF.Draw, "d", "draw").
	Add(usecase.OsmosisInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.OsmosisInteractorIF.AutoComplete, "ac", "autocomplete").
	Add(usecase.OsmosisInteractorIF.Undo, "u", "undo").
	Add(usecase.OsmosisInteractorIF.Hint, "h", "hint").
	Add(usecase.OsmosisInteractorIF.ActionLog, "log", "l")

// osmosisArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var osmosisArgfulCommands = []string{"m", "move"}

// OsmosisCuiController オズモシスCUIコントローラークラス
type OsmosisCuiController struct {
	oi usecase.OsmosisInteractorIF
}

// NewOsmosisCuiController コンストラクタ
func NewOsmosisCuiController(oi usecase.OsmosisInteractorIF) *OsmosisCuiController {
	return &OsmosisCuiController{oi: oi}
}

// Exec コマンド実行
func (c *OsmosisCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.oi.Reset() },
		append(osmosisNoArgCommands.Names(), osmosisArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := osmosisNoArgCommands.Lookup(cmd); ok {
				return fn(c.oi), true
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

// handleMove 移動コマンドを処理する。
//
// 受理する書式:
//
//	m w f <fIdx>          ウェイスト → ファンデーション fIdx 段目
//	m r <rIdx> f <fIdx>   リザーブ rIdx 列 → ファンデーション fIdx 段目
func (c *OsmosisCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("osmosis.promptSourceZone"), "m {0}")
	}
	from := args[0]
	switch from {
	case "w":
		return c.handleMoveFromWaste(args[1:])
	case "r":
		return c.handleMoveFromReserve(args[1:])
	default:
		return invalidArg("osmosis.invalidFromZone", "val", from)
	}
}

func (c *OsmosisCuiController) handleMoveFromWaste(args []string) string {
	// args: ["f", "<fIdx>"]
	if len(args) < 1 || args[0] != "f" {
		return i18n.MarkError(i18n.T("osmosis.moveUsage"))
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("osmosis.promptFoundation"), "m w f {0}")
	}
	fIdx, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[1])
	}
	return c.oi.MoveWasteToFoundation(fIdx)
}

func (c *OsmosisCuiController) handleMoveFromReserve(args []string) string {
	// args: ["<rIdx>", "f", "<fIdx>"]
	if len(args) < 1 {
		return cuiutil.PromptRequest(i18n.T("osmosis.promptReserve"), "m r {0}")
	}
	rIdx, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}
	if len(args) < 2 || args[1] != "f" {
		return i18n.MarkError(i18n.T("osmosis.moveUsage"))
	}
	if len(args) < 3 {
		return cuiutil.PromptRequest(i18n.T("osmosis.promptFoundation"), "m r "+args[0]+" f {0}")
	}
	fIdx, err := strconv.Atoi(args[2])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[2])
	}
	return c.oi.MoveReserveToFoundation(rIdx, fIdx)
}
