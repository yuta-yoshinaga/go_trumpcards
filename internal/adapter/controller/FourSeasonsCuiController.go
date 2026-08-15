//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// fourSeasonsNoArgCommands maps no-arg CUI commands to FourSeasons interactor methods.
var fourSeasonsNoArgCommands = cuiutil.NewCommandMap[usecase.FourSeasonsInteractorIF]().
	Add(usecase.FourSeasonsInteractorIF.Draw, "d", "draw").
	Add(usecase.FourSeasonsInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.FourSeasonsInteractorIF.AutoComplete, "ac", "autocomplete").
	Add(usecase.FourSeasonsInteractorIF.Undo, "u", "undo").
	Add(usecase.FourSeasonsInteractorIF.Hint, "h", "hint").
	Add(usecase.FourSeasonsInteractorIF.ActionLog, "log", "l")

// fourSeasonsArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var fourSeasonsArgfulCommands = []string{"m", "move"}

// FourSeasonsCuiController フォーシーズンズCUIコントローラークラス
type FourSeasonsCuiController struct {
	ci usecase.FourSeasonsInteractorIF
}

// NewFourSeasonsCuiController コンストラクタ
func NewFourSeasonsCuiController(ci usecase.FourSeasonsInteractorIF) *FourSeasonsCuiController {
	return &FourSeasonsCuiController{ci: ci}
}

// Exec コマンド実行
func (c *FourSeasonsCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ci.Reset() },
		append(fourSeasonsNoArgCommands.Names(), fourSeasonsArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := fourSeasonsNoArgCommands.Lookup(cmd); ok {
				return fn(c.ci), true
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

// handleMove は `m <from> <to>` を捌く。
//
//	m w f <fIdx>   ウェイスト → ファンデーション
//	m w t <col>    ウェイスト → タブロー
//	m t <col> f <fIdx>  タブロー → ファンデーション
//	m t <col> t <col>   タブロー → タブロー
//
// **ファンデーションは必ず番号を要求する。** 四隅のどこが開くかはベースランクで
// 開いた順に決まるので、行き先を1つに決め打ちできない。
func (c *FourSeasonsCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("fourseasons.promptSourceZone"), "m {0}")
	}
	switch args[0] {
	case "w":
		return c.handleFromWaste(args)
	case "t":
		return c.handleFromTableau(args)
	default:
		return invalidArg("fourseasons.invalidFromZone", "val", args[0])
	}
}

func (c *FourSeasonsCuiController) handleFromWaste(args []string) string {
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("fourseasons.promptToZone"), "m w {0}")
	}
	dest := args[1]
	if dest != "f" && dest != "t" {
		return invalidArg("fourseasons.invalidToZone", "val", dest)
	}
	if len(args) < 3 {
		return cuiutil.PromptRequest(i18n.T("fourseasons.promptIndex"), fmt.Sprintf("m w %s {0}", dest))
	}
	idx, err := strconv.Atoi(args[2])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[2])
	}
	if dest == "f" {
		return c.ci.MoveWasteToFoundation(idx)
	}
	return c.ci.MoveWasteToTableau(idx)
}

func (c *FourSeasonsCuiController) handleFromTableau(args []string) string {
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	fromCol, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[1])
	}
	if len(args) < 3 {
		return cuiutil.PromptRequest(i18n.T("fourseasons.promptToZone"), fmt.Sprintf("m t %d {0}", fromCol))
	}
	dest := args[2]
	if dest != "f" && dest != "t" {
		return invalidArg("fourseasons.invalidToZone", "val", dest)
	}
	if len(args) < 4 {
		return cuiutil.PromptRequest(i18n.T("fourseasons.promptIndex"), fmt.Sprintf("m t %d %s {0}", fromCol, dest))
	}
	toIdx, err := strconv.Atoi(args[3])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[3])
	}
	if dest == "f" {
		return c.ci.MoveTableauToFoundation(fromCol, toIdx)
	}
	return c.ci.MoveTableauToTableau(fromCol, toIdx)
}
