package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ScorpionCuiController スコーピオンCUIコントローラークラス
type ScorpionCuiController struct {
	si usecase.ScorpionInteractorIF
}

// NewScorpionCuiController コンストラクタ
func NewScorpionCuiController(si usecase.ScorpionInteractorIF) *ScorpionCuiController {
	return &ScorpionCuiController{si: si}
}

// Exec コマンド実行
func (c *ScorpionCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.si.Reset() },
		[]string{"d", "deal", "m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "deal":
				return c.si.Deal(), true
			case "m", "move":
				return c.handleMove(args), true
			case "g", "giveup":
				return c.si.GiveUp(), true
			case "ac", "autocomplete":
				return c.si.AutoComplete(), true
			case "u", "undo":
				return c.si.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.si.Hint, c.si.ActionLog)
			}
		},
	)
}

// handleMove 移動コマンドを処理
// Format: m t <fromCol> <cardIdx> t <toCol>
// Shorthand: m <fromCol> <toCol> (top card) or m <fromCol> <cardIdx> <toCol>
func (c *ScorpionCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("scorpion.promptFromColumn"), "m {0}")
	}
	if _, err := strconv.Atoi(args[0]); err == nil {
		return c.handleMoveShorthand(args)
	}
	if args[0] != "t" {
		return i18n.T("scorpion.moveUsage")
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("scorpion.promptFromColumn"), "m t {0}")
	}
	fromCol, err := strconv.Atoi(args[1])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[1])
	}
	if len(args) < 3 {
		return cuiutil.PromptRequest(i18n.T("scorpion.promptCardIndex"), fmt.Sprintf("m t %d {0} t", fromCol))
	}
	if len(args) < 5 || args[3] != "t" {
		if len(args) == 3 || (len(args) == 4 && args[3] == "t") {
			return cuiutil.PromptRequest(i18n.T("scorpion.promptToColumn"), fmt.Sprintf("m t %d %s t {0}", fromCol, args[2]))
		}
		return i18n.T("scorpion.moveUsage")
	}
	cardIdx, err := strconv.Atoi(args[2])
	if err != nil {
		return i18n.Tf("invalidCardIndex", "val", args[2])
	}
	toCol, err := strconv.Atoi(args[4])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[4])
	}
	return c.si.MoveTableauToTableau(fromCol, cardIdx, toCol)
}

func (c *ScorpionCuiController) handleMoveShorthand(args []string) string {
	fromCol, _ := strconv.Atoi(args[0])
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("scorpion.promptToColumn"), fmt.Sprintf("m %s {0}", args[0]))
	}
	if len(args) == 2 {
		toCol, err := strconv.Atoi(args[1])
		if err != nil {
			return i18n.Tf("invalidColumn", "val", args[1])
		}
		return c.si.MoveTableauToTableau(fromCol, -1, toCol)
	}
	cardIdx, err := strconv.Atoi(args[1])
	if err != nil {
		return i18n.Tf("invalidCardIndex", "val", args[1])
	}
	toCol, err := strconv.Atoi(args[2])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[2])
	}
	return c.si.MoveTableauToTableau(fromCol, cardIdx, toCol)
}
