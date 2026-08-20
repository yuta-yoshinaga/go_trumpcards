//go:build !js || !wasm || extra

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// agnesNoArgCommands maps no-arg CUI commands to Agnes interactor methods.
var agnesNoArgCommands = cuiutil.NewCommandMap[usecase.AgnesInteractorIF]().
	Add(usecase.AgnesInteractorIF.DealStock, "d", "deal").
	Add(usecase.AgnesInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.AgnesInteractorIF.Undo, "u", "undo").
	Add(usecase.AgnesInteractorIF.Hint, "h", "hint").
	Add(usecase.AgnesInteractorIF.ActionLog, "log", "l")

// agnesArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var agnesArgfulCommands = []string{"m", "move"}

// AgnesCuiController アグネス・ソレルCUIコントローラークラス
type AgnesCuiController struct {
	ci usecase.AgnesInteractorIF
}

// NewAgnesCuiController コンストラクタ
func NewAgnesCuiController(ci usecase.AgnesInteractorIF) *AgnesCuiController {
	return &AgnesCuiController{ci: ci}
}

// Exec コマンド実行
func (c *AgnesCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ci.Reset() },
		append(agnesNoArgCommands.Names(), agnesArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := agnesNoArgCommands.Lookup(cmd); ok {
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

// handleMove 移動コマンドを処理
//
// 受け付ける形式:
//
//	m <fromCol> <toCol>  : タブロー間移動 (末尾カード)
//	m <fromCol> f        : タブロー→ファンデーション
func (c *AgnesCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m {0}")
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("agnes.promptToZone"), fmt.Sprintf("m %s {0}", args[0]))
	}
	if args[1] == "f" {
		return c.ci.MoveTableauToFoundation(fromCol)
	}
	toCol, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[1])
	}
	return c.ci.MoveTableauToTableau(fromCol, -1, toCol)
}
