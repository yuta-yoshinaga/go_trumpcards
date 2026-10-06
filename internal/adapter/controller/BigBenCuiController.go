//go:build !js || !wasm || extra3

package controller

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BigBenCuiController ビッグ・ベン CUI コントローラークラス
type BigBenCuiController struct {
	gi usecase.BigBenInteractorIF
}

// NewBigBenCuiController コンストラクタ
func NewBigBenCuiController(gi usecase.BigBenInteractorIF) *BigBenCuiController {
	return &BigBenCuiController{gi: gi}
}

// Exec コマンド実行
func (c *BigBenCuiController) Exec(command string) string {
	return execSolitaireCui(command, solitaireCuiFns{
		reset:        c.gi.Reset,
		move:         c.handleMove,
		giveUp:       c.gi.GiveUp,
		autoComplete: c.gi.AutoComplete,
		undo:         c.gi.Undo,
		hint:         c.gi.Hint,
		actionLog:    c.gi.ActionLog,
		extraCommands: map[string]func([]string) string{
			"d":    func([]string) string { return c.gi.Deal() },
			"deal": func([]string) string { return c.gi.Deal() },
		},
	})
}

// handleMove 移動コマンドを処理。supported syntax:
//
//	m <col> f <fIdx>      - tableau top to clock face <fIdx> (0..11)
//	m <fromCol> <toCol>   - tableau to tableau
//
// 文字盤は 12 個あって同スートの札が複数に載りうるため、送り先の番号は省略
// できない。
func (c *BigBenCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m {0}")
	}
	fromCol, msg, ok := cuiutil.ParseIntArgKeys(args, "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
	if !ok {
		return msg
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("bigben.promptToZone"), fmt.Sprintf("m %s {0}", args[0]))
	}
	if args[1] == "f" {
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("bigben.promptFaceIdx"),
				fmt.Sprintf("m %s f {0}", args[0]))
		}
		fIdx, msg, ok := cuiutil.ParseIntArgKeys(args[2:], "", "bigben.invalidFaceIdx", cuiutil.NoMin, cuiutil.NoMax)
		if !ok {
			return msg
		}
		return c.gi.MoveTableauToFoundation(fromCol, fIdx)
	}
	toCol, msg, ok := cuiutil.ParseIntArgKeys(args[1:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
	if !ok {
		return msg
	}
	return c.gi.MoveTableauToTableau(fromCol, toCol)
}
