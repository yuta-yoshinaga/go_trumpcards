//go:build !js || !wasm || solo

package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// spiderNoArgCommands maps no-arg CUI commands to Spider interactor methods.
var spiderNoArgCommands = cuiutil.NewCommandMap[usecase.SpiderInteractorIF]().
	Add(usecase.SpiderInteractorIF.Deal, "d", "deal").
	Add(usecase.SpiderInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.SpiderInteractorIF.AutoComplete, "ac", "autocomplete").
	Add(usecase.SpiderInteractorIF.Undo, "u", "undo").
	Add(usecase.SpiderInteractorIF.Hint, "h", "hint").
	Add(usecase.SpiderInteractorIF.ActionLog, "log", "l")

// spiderArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var spiderArgfulCommands = []string{"m", "move"}

// SpiderCuiController スパイダーソリティアCUIコントローラークラス
type SpiderCuiController struct {
	si usecase.SpiderInteractorIF
}

// NewSpiderCuiController コンストラクタ
func NewSpiderCuiController(si usecase.SpiderInteractorIF) *SpiderCuiController {
	return &SpiderCuiController{si: si}
}

// Exec コマンド実行
func (c *SpiderCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(args []string) string {
			if len(args) > 0 {
				n, err := strconv.Atoi(args[0])
				if err == nil && (n == 1 || n == 2 || n == 4) {
					return c.si.ResetWithConfig(domain.SpiderConfig{Difficulty: domain.SpiderDifficulty(n)})
				}
			}
			return c.si.Reset()
		},
		append(spiderNoArgCommands.Names(), spiderArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := spiderNoArgCommands.Lookup(cmd); ok {
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
func (c *SpiderCuiController) handleMove(args []string) string {
	// Wizard-style prompts for missing arguments
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m {0}")
	}
	// Shorthand: numeric first arg means tableau shorthand
	if _, err := strconv.Atoi(args[0]); err == nil {
		return c.handleMoveShorthand(args)
	}
	if args[0] != "t" {
		return i18n.T("spider.moveUsage")
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
		return i18n.T("spider.moveUsage")
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

func (c *SpiderCuiController) handleMoveShorthand(args []string) string {
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
