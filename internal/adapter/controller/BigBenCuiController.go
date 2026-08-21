//go:build !js || !wasm || extra3

package controller

import (
	"fmt"
	"strconv"

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
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.gi.Reset()
		},
		[]string{"d", "deal", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "deal":
				return c.gi.Deal(), true
			case "m", "move":
				return c.handleMove(args), true
			case "g", "giveup":
				return c.gi.GiveUp(), true
			case "ac", "autocomplete":
				return c.gi.AutoComplete(), true
			case "u", "undo":
				return c.gi.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.gi.Hint, c.gi.ActionLog)
			}
		},
	)
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
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("bigben.promptToZone"), fmt.Sprintf("m %s {0}", args[0]))
	}
	if args[1] == "f" {
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("bigben.promptFaceIdx"),
				fmt.Sprintf("m %s f {0}", args[0]))
		}
		fIdx, err := strconv.Atoi(args[2])
		if err != nil {
			return invalidArg("bigben.invalidFaceIdx", "val", args[2])
		}
		return c.gi.MoveTableauToFoundation(fromCol, fIdx)
	}
	toCol, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[1])
	}
	return c.gi.MoveTableauToTableau(fromCol, toCol)
}
