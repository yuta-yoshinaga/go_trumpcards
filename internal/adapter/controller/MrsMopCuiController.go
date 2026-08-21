//go:build !js || !wasm || extra4

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// mrsMopNoArgCommands maps no-arg CUI commands to MrsMop interactor methods.
var mrsMopNoArgCommands = cuiutil.NewCommandMap[usecase.MrsMopInteractorIF]().
	Add(usecase.MrsMopInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.MrsMopInteractorIF.AutoComplete, "ac", "autocomplete").
	Add(usecase.MrsMopInteractorIF.Undo, "u", "undo").
	Add(usecase.MrsMopInteractorIF.Hint, "h", "hint").
	Add(usecase.MrsMopInteractorIF.ActionLog, "log", "l")

// mrsMopArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var mrsMopArgfulCommands = []string{"m", "move"}

// MrsMopCuiController ミセス・モップソリティアCUIコントローラークラス
type MrsMopCuiController struct {
	si usecase.MrsMopInteractorIF
}

// NewMrsMopCuiController コンストラクタ
func NewMrsMopCuiController(si usecase.MrsMopInteractorIF) *MrsMopCuiController {
	return &MrsMopCuiController{si: si}
}

// Exec コマンド実行
func (c *MrsMopCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(args []string) string {
			if len(args) > 0 {
				n, err := strconv.Atoi(args[0])
				if err == nil && (n == 1 || n == 2 || n == 4) {
					return c.si.ResetWithConfig(domain.MrsMopConfig{Difficulty: domain.MrsMopDifficulty(n)})
				}
			}
			return c.si.Reset()
		},
		append(mrsMopNoArgCommands.Names(), mrsMopArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := mrsMopNoArgCommands.Lookup(cmd); ok {
				return fn(c.si), true
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
// Format: m t <fromCol> <cardIdx> t <toCol>
// Shorthand: m <fromCol> <toCol> (top card) or m <fromCol> <cardIdx> <toCol>
func (c *MrsMopCuiController) handleMove(args []string) string {
	// Wizard-style prompts for missing arguments
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m {0}")
	}
	// Shorthand: numeric first arg means tableau shorthand
	if _, err := strconv.Atoi(args[0]); err == nil {
		return c.handleMoveShorthand(args)
	}
	if args[0] != "t" {
		return i18n.MarkError(i18n.T("mrsmop.moveUsage"))
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m t {0}")
	}
	fromCol, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[1])
	}
	if len(args) < 3 {
		return cuiutil.PromptRequest(i18n.T("promptCardIndex"), fmt.Sprintf("m t %d {0} t", fromCol))
	}
	if len(args) < 5 || args[3] != "t" {
		if len(args) == 3 || (len(args) == 4 && args[3] == "t") {
			return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m t %d %s t {0}", fromCol, args[2]))
		}
		return i18n.MarkError(i18n.T("mrsmop.moveUsage"))
	}
	cardIdx, err := strconv.Atoi(args[2])
	if err != nil {
		return invalidArg("invalidCardIndex", "val", args[2])
	}
	toCol, err := strconv.Atoi(args[4])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[4])
	}
	return c.si.MoveTableauToTableau(fromCol, cardIdx, toCol)
}

func (c *MrsMopCuiController) handleMoveShorthand(args []string) string {
	fromCol, _ := strconv.Atoi(args[0])
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("promptToColumn"), fmt.Sprintf("m %s {0}", args[0]))
	}
	// m <fromCol> <toCol> — top card move
	if len(args) == 2 {
		toCol, err := strconv.Atoi(args[1])
		if err != nil {
			return invalidArg("invalidColumn", "val", args[1])
		}
		return c.si.MoveTableauToTableau(fromCol, -1, toCol)
	}
	// m <fromCol> <cardIdx> <toCol>
	cardIdx, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidCardIndex", "val", args[1])
	}
	toCol, err := strconv.Atoi(args[2])
	if err != nil {
		return invalidArg("invalidColumn", "val", args[2])
	}
	return c.si.MoveTableauToTableau(fromCol, cardIdx, toCol)
}
