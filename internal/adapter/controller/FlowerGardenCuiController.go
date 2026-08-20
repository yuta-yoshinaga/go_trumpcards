//go:build !js || !wasm || extra4

package controller

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FlowerGardenCuiController Flower Garden CUI コントローラークラス
type FlowerGardenCuiController struct {
	bi usecase.FlowerGardenInteractorIF
}

// NewFlowerGardenCuiController コンストラクタ
func NewFlowerGardenCuiController(bi usecase.FlowerGardenInteractorIF) *FlowerGardenCuiController {
	return &FlowerGardenCuiController{bi: bi}
}

// Exec コマンド実行
func (c *FlowerGardenCuiController) Exec(command string) string {
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

// handleMove 移動コマンドを処理。
// Flower Garden supported syntax:
//
//	m <fromCol> <toCol>   - move end card between flower-bed fans
//	m <fromCol> f         - move end tableau card to foundation
//	m r<idx> <toCol>      - move a reserve card to a tableau fan
//	m r<idx> f            - move a reserve card to foundation
func (c *FlowerGardenCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m {0}")
	}
	from := args[0]
	// Reserve source: prefixed with "r" (e.g. r0).
	if strings.HasPrefix(from, "r") {
		reserveIdx, err := strconv.Atoi(strings.TrimPrefix(from, "r"))
		if err != nil {
			return invalidArg("invalidColumn", "val", from)
		}
		if len(args) < 2 {
			return cuiutil.PromptRequest(i18n.T("flowergarden.promptToZone"), fmt.Sprintf("m %s {0}", from))
		}
		if args[1] == "f" {
			return c.bi.MoveReserveToFoundation(reserveIdx)
		}
		toCol, err := strconv.Atoi(args[1])
		if err != nil {
			return invalidArg("invalidColumn", "val", args[1])
		}
		return c.bi.MoveReserveToTableau(reserveIdx, toCol)
	}

	fromCol, err := strconv.Atoi(from)
	if err != nil {
		return invalidArg("invalidColumn", "val", from)
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("flowergarden.promptToZone"), fmt.Sprintf("m %s {0}", from))
	}
	if args[1] == "f" {
		return c.bi.MoveTableauToFoundation(fromCol)
	}
	toCol, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[1])
	}
	return c.bi.MoveTableauToTableau(fromCol, -1, toCol)
}
