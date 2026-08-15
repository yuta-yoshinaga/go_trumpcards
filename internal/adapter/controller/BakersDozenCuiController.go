//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BakersDozenCuiController ベーカーズダズンCUIコントローラークラス
type BakersDozenCuiController struct {
	bi usecase.BakersDozenInteractorIF
}

// NewBakersDozenCuiController コンストラクタ
func NewBakersDozenCuiController(bi usecase.BakersDozenInteractorIF) *BakersDozenCuiController {
	return &BakersDozenCuiController{bi: bi}
}

// Exec コマンド実行
func (c *BakersDozenCuiController) Exec(command string) string {
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
// Baker's Dozen has no waste/stock; supported syntax:
//
//	m <fromCol> <toCol>   - move top card between tableau columns
//	m <fromCol> f         - move top card to foundation
func (c *BakersDozenCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m {0}")
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("bakersdozen.promptToZone"), fmt.Sprintf("m %s {0}", args[0]))
	}
	if args[1] == "f" {
		return c.bi.MoveTableauToFoundation(fromCol)
	}
	toCol, err := strconv.Atoi(args[1])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[1])
	}
	return c.bi.MoveTableauToTableau(fromCol, -1, toCol)
}
