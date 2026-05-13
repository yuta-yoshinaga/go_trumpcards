package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BeleagueredCastleCuiController Beleaguered Castle CUI コントローラークラス
type BeleagueredCastleCuiController struct {
	bi usecase.BeleagueredCastleInteractorIF
}

// NewBeleagueredCastleCuiController コンストラクタ
func NewBeleagueredCastleCuiController(bi usecase.BeleagueredCastleInteractorIF) *BeleagueredCastleCuiController {
	return &BeleagueredCastleCuiController{bi: bi}
}

// Exec コマンド実行
func (c *BeleagueredCastleCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.bi.Reset()
		},
		[]string{"m", "move", "g", "giveup", "h", "hint", "ac", "autocomplete", "log", "l", "u", "undo"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "m", "move":
				return c.handleMove(args), true
			case "g", "giveup":
				return c.bi.GiveUp(), true
			case "ac", "autocomplete":
				return c.bi.AutoComplete(), true
			case "u", "undo":
				return c.bi.Undo(), true
			default:
				return handleCuiHintAndLog(cmd, c.bi.Hint, c.bi.ActionLog)
			}
		},
	)
}

// handleMove 移動コマンドを処理
// Beleaguered Castle has no waste/stock; supported syntax:
//
//	m <fromCol> <toCol>   - move top card between tableau columns
//	m <fromCol> f         - move top card to foundation
func (c *BeleagueredCastleCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m {0}")
	}
	fromCol, err := strconv.Atoi(args[0])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[0])
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("beleagueredcastle.promptToZone"), fmt.Sprintf("m %s {0}", args[0]))
	}
	if args[1] == "f" {
		return c.bi.MoveTableauToFoundation(fromCol)
	}
	toCol, err := strconv.Atoi(args[1])
	if err != nil {
		return i18n.Tf("invalidColumn", "val", args[1])
	}
	return c.bi.MoveTableauToTableau(fromCol, -1, toCol)
}
