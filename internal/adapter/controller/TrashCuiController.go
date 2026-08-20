package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TrashCuiController トラッシュCUIコントローラークラス
type TrashCuiController struct {
	ti usecase.TrashInteractorIF
}

// NewTrashCuiController コンストラクタ
func NewTrashCuiController(ti usecase.TrashInteractorIF) *TrashCuiController {
	return &TrashCuiController{ti: ti}
}

// Exec コマンド実行
func (c *TrashCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ti.Reset() },
		[]string{"d", "draw", "p", "place", "cpu", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				return c.ti.Draw(), true
			case "p", "place":
				return c.handlePlace(args), true
			case "cpu":
				return c.ti.CpuStep(), true
			case "h", "hint":
				return c.ti.Hint(), true
			default:
				result, handled := handleCuiLog(cmd, c.ti.ActionLog)
				return result, handled
			}
		},
	)
}

// handlePlace 位置指定によるワイルド配置
// 形式: p <pos>
func (c *TrashCuiController) handlePlace(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("trash.promptPosition"), "p {0}")
	}
	pos, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidIndex", "val", args[0])
	}
	return c.ti.PlaceWild(pos)
}
