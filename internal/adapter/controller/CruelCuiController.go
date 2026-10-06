//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CruelCuiController クルーエルCUIコントローラークラス
type CruelCuiController struct {
	ci usecase.CruelInteractorIF
}

// NewCruelCuiController コンストラクタ
func NewCruelCuiController(ci usecase.CruelInteractorIF) *CruelCuiController {
	return &CruelCuiController{ci: ci}
}

// Exec コマンド実行
func (c *CruelCuiController) Exec(command string) string {
	return execSolitaireCui(command, solitaireCuiFns{
		reset:        c.ci.Reset,
		move:         c.handleMove,
		giveUp:       c.ci.GiveUp,
		autoComplete: c.ci.AutoComplete,
		undo:         c.ci.Undo,
		hint:         c.ci.Hint,
		actionLog:    c.ci.ActionLog,
		extraCommands: map[string]func([]string) string{
			"s":     func([]string) string { return c.ci.Shift() },
			"shift": func([]string) string { return c.ci.Shift() },
		},
	})
}

// handleMove 移動コマンドを処理（Cruel は最上段のみ移動できるので cardIndex なし）
func (c *CruelCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("cruel.promptSourceColumn"), "m {0}")
	}

	fromCol, msg, ok := cuiutil.ParseIntArgKeys(args, "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
	if !ok {
		return msg
	}

	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("cruel.promptToZone"), fmt.Sprintf("m %d {0}", fromCol))
	}

	dest := args[1]
	if dest == "f" {
		return c.ci.MoveTableauToFoundation(fromCol)
	}
	toCol, err := strconv.Atoi(dest)
	if err != nil {
		return i18n.MarkError(i18n.T("cruel.moveUsage"))
	}
	return c.ci.MoveTableauToTableau(fromCol, toCol)
}
