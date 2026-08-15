//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// crescentNoArgCommands 引数なし CUI コマンドのマップ。
// 注: "r" は execCuiCommand が reset として横取りするため、
// Redeal の短縮形には "rd" を使う。
var crescentNoArgCommands = cuiutil.NewCommandMap[usecase.CrescentInteractorIF]().
	Add(usecase.CrescentInteractorIF.Redeal, "rd", "redeal").
	Add(usecase.CrescentInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.CrescentInteractorIF.AutoComplete, "ac", "autocomplete").
	Add(usecase.CrescentInteractorIF.Undo, "u", "undo").
	Add(usecase.CrescentInteractorIF.Hint, "h", "hint").
	Add(usecase.CrescentInteractorIF.ActionLog, "log", "l")

// crescentArgfulCommands 引数あり CUI コマンドの別名。
var crescentArgfulCommands = []string{"m", "move"}

// CrescentCuiController クレセント・ソリティアの CUI コントローラー。
type CrescentCuiController struct {
	ci usecase.CrescentInteractorIF
}

// NewCrescentCuiController コンストラクタ。
func NewCrescentCuiController(ci usecase.CrescentInteractorIF) *CrescentCuiController {
	return &CrescentCuiController{ci: ci}
}

// Exec CUI コマンドを実行する。
func (c *CrescentCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.ci.Reset()
		},
		append(crescentNoArgCommands.Names(), crescentArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := crescentNoArgCommands.Lookup(cmd); ok {
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

// handleMove 移動コマンドを処理する。
//
// 受け付ける形式:
//
//	m <fromCol> <toCol>            タブロー間の移動 (簡略形)
//	m t <fromCol> t <toCol>        タブロー間の移動 (明示形)
//	m t <fromCol> f <foundationId> タブロー→ファンデーション
func (c *CrescentCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("crescent.promptFromColumn"), "m {0}")
	}
	// Shorthand: m <fromCol> <toCol>
	if _, err := strconv.Atoi(args[0]); err == nil {
		return c.handleMoveShorthand(args)
	}
	if args[0] != "t" {
		return invalidArg("crescent.invalidFromZone", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("crescent.promptFromColumn"), "m t {0}")
	}
	fromCol, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[1])
	}
	if len(args) < 3 {
		return cuiutil.PromptRequest(i18n.T("crescent.promptToZone"), fmt.Sprintf("m t %s {0}", args[1]))
	}
	switch args[2] {
	case "t":
		if len(args) < 4 {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m t %s t {0}", args[1]))
		}
		toCol, err := strconv.Atoi(args[3])
		if err != nil {
			return invalidArg("invalidColumn", "val", args[3])
		}
		return c.ci.MoveTableauToTableau(fromCol, toCol)
	case "f":
		if len(args) < 4 {
			return cuiutil.PromptRequest(i18n.T("crescent.promptFoundationId"), fmt.Sprintf("m t %s f {0}", args[1]))
		}
		fIdx, err := strconv.Atoi(args[3])
		if err != nil {
			return invalidArg("crescent.invalidFoundationId", "val", args[3])
		}
		return c.ci.MoveTableauToFoundation(fromCol, fIdx)
	default:
		return i18n.T("crescent.moveUsage")
	}
}

func (c *CrescentCuiController) handleMoveShorthand(args []string) string {
	fromCol, _ := strconv.Atoi(args[0])
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m %s {0}", args[0]))
	}
	toCol, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[1])
	}
	return c.ci.MoveTableauToTableau(fromCol, toCol)
}
