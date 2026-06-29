//go:build !js || !wasm || extra

package controller

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KingAlbertCuiController King Albert CUI コントローラークラス
type KingAlbertCuiController struct {
	bi usecase.KingAlbertInteractorIF
}

// NewKingAlbertCuiController コンストラクタ
func NewKingAlbertCuiController(bi usecase.KingAlbertInteractorIF) *KingAlbertCuiController {
	return &KingAlbertCuiController{bi: bi}
}

// Exec コマンド実行
func (c *KingAlbertCuiController) Exec(command string) string {
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

// handleMove 移動コマンドを処理。
// King Albert supported syntax:
//
//	m <fromCol> <toCol>   - move bottom card between tableau columns
//	m <fromCol> f         - move bottom tableau card to foundation
//	m r<idx> <toCol>      - move a reserve card to a tableau column
//	m r<idx> f            - move a reserve card to foundation
func (c *KingAlbertCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("promptFromColumn"), "m {0}")
	}
	from := args[0]
	// Reserve source: prefixed with "r" (e.g. r0).
	if strings.HasPrefix(from, "r") {
		reserveIdx, err := strconv.Atoi(strings.TrimPrefix(from, "r"))
		if err != nil {
			return i18n.Tf("invalidColumn", "val", from)
		}
		if len(args) < 2 {
			return cuiutil.PromptRequest(i18n.T("kingalbert.promptToZone"), fmt.Sprintf("m %s {0}", from))
		}
		if args[1] == "f" {
			return c.bi.MoveReserveToFoundation(reserveIdx)
		}
		toCol, err := strconv.Atoi(args[1])
		if err != nil {
			return i18n.Tf("invalidColumn", "val", args[1])
		}
		return c.bi.MoveReserveToTableau(reserveIdx, toCol)
	}

	fromCol, err := strconv.Atoi(from)
	if err != nil {
		return i18n.Tf("invalidColumn", "val", from)
	}
	if len(args) < 2 {
		return cuiutil.PromptRequest(i18n.T("kingalbert.promptToZone"), fmt.Sprintf("m %s {0}", from))
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
