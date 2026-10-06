//go:build !js || !wasm || solo

package controller

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FortressCuiController Fortress CUI コントローラークラス
type FortressCuiController struct {
	bi usecase.FortressInteractorIF
}

// NewFortressCuiController コンストラクタ
func NewFortressCuiController(bi usecase.FortressInteractorIF) *FortressCuiController {
	return &FortressCuiController{bi: bi}
}

// Exec コマンド実行
func (c *FortressCuiController) Exec(command string) string {
	return execSolitaireCui(command, solitaireCuiFns{
		reset:        c.bi.Reset,
		move:         c.handleMove,
		giveUp:       c.bi.GiveUp,
		autoComplete: c.bi.AutoComplete,
		undo:         c.bi.Undo,
		hint:         c.bi.Hint,
		actionLog:    c.bi.ActionLog,
	})
}

// handleMove 移動コマンドを処理
// Fortress has no waste/stock; supported syntax:
//
//	m <fromCol> <toCol>   - move top card between tableau columns
//	m <fromCol> f         - move top card to foundation
func (c *FortressCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m {0}")
	}
	fromCol, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args, "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
	if !parseOK {
		return parseMsg
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("fortress.promptToZone"), fmt.Sprintf("m %s {0}", args[0]))
	}
	if args[1] == "f" {
		return c.bi.MoveTableauToFoundation(fromCol)
	}
	toCol, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[1:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
	if !parseOK {
		return parseMsg
	}
	return c.bi.MoveTableauToTableau(fromCol, -1, toCol)
}
