//go:build !js || !wasm || extra2

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BisleyCuiController ビズリー CUI コントローラークラス
type BisleyCuiController struct {
	bi usecase.BisleyInteractorIF
}

// NewBisleyCuiController コンストラクタ
func NewBisleyCuiController(bi usecase.BisleyInteractorIF) *BisleyCuiController {
	return &BisleyCuiController{bi: bi}
}

// Exec コマンド実行
func (c *BisleyCuiController) Exec(command string) string {
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
// Bisley has no stock or waste, and two foundation sets; supported syntax:
//
//	m <fromCol> a         - move top card to the ascending (ace) foundation
//	m <fromCol> k         - move top card to the descending (king) foundation
//	m <fromCol> <toCol>   - move top card between tableau columns
func (c *BisleyCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m {0}")
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("bisley.promptToZone"), fmt.Sprintf("m %s {0}", args[0]))
	}
	switch args[1] {
	case "a":
		return c.bi.MoveTableauToAceFoundation(fromCol)
	case "k":
		return c.bi.MoveTableauToKingFoundation(fromCol)
	}
	toCol, err := strconv.Atoi(args[1])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[1])
	}
	return c.bi.MoveTableauToTableau(fromCol, toCol)
}
