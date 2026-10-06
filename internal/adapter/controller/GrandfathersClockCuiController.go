//go:build !js || !wasm || extra2

package controller

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GrandfathersClockCuiController グランドファーザーズ・クロック CUI コントローラークラス
type GrandfathersClockCuiController struct {
	gi usecase.GrandfathersClockInteractorIF
}

// NewGrandfathersClockCuiController コンストラクタ
func NewGrandfathersClockCuiController(gi usecase.GrandfathersClockInteractorIF) *GrandfathersClockCuiController {
	return &GrandfathersClockCuiController{gi: gi}
}

// Exec コマンド実行
func (c *GrandfathersClockCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.gi.Reset()
		},
		[]string{"m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo", "redo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "m", "move":
				return c.handleMove(args), true
			case "g", "giveup":
				return c.gi.GiveUp(), true
			case "ac", "autocomplete":
				return c.gi.AutoComplete(), true
			case "u", "undo":
				return c.gi.Undo(), true
			case "redo":
				return c.gi.Redo(), true
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
func (c *GrandfathersClockCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m {0}")
	}
	fromCol, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[0:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
	if !parseOK {
		return parseMsg
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("grandfathersclock.promptToZone"), fmt.Sprintf("m %s {0}", args[0]))
	}
	if args[1] == "f" {
		if len(args) < 3 {
			return cuiutil.PromptRequest(i18n.T("grandfathersclock.promptFaceIdx"),
				fmt.Sprintf("m %s f {0}", args[0]))
		}
		fIdx, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[2:], "", "grandfathersclock.invalidFaceIdx", cuiutil.NoMin, cuiutil.NoMax)
		if !parseOK {
			return parseMsg
		}
		return c.gi.MoveTableauToFoundation(fromCol, fIdx)
	}
	toCol, parseMsg, parseOK := cuiutil.ParseIntArgKeys(args[1:], "", "invalidColumn", cuiutil.NoMin, cuiutil.NoMax)
	if !parseOK {
		return parseMsg
	}
	return c.gi.MoveTableauToTableau(fromCol, toCol)
}
