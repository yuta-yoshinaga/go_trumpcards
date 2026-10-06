//go:build !js || !wasm || solo

package controller

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CitadelCuiController Citadel CUI コントローラークラス
type CitadelCuiController struct {
	bi usecase.CitadelInteractorIF
}

// NewCitadelCuiController コンストラクタ
func NewCitadelCuiController(bi usecase.CitadelInteractorIF) *CitadelCuiController {
	return &CitadelCuiController{bi: bi}
}

// Exec コマンド実行
func (c *CitadelCuiController) Exec(command string) string {
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
// Citadel has no waste/stock; supported syntax:
//
//	m <fromCol> <toCol>   - move top card between tableau columns
//	m <fromCol> f         - move top card to foundation
func (c *CitadelCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m {0}")
	}
	fromCol, msg, ok := cuiutil.ParseIntArgKeys(args, "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
	if !ok {
		return msg
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("citadel.promptToZone"), fmt.Sprintf("m %s {0}", args[0]))
	}
	if args[1] == "f" {
		return c.bi.MoveTableauToFoundation(fromCol)
	}
	toCol, msg, ok := cuiutil.ParseIntArgKeys(args[1:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
	if !ok {
		return msg
	}
	return c.bi.MoveTableauToTableau(fromCol, -1, toCol)
}
