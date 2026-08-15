//go:build !js || !wasm || extra

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// sultanNoArgCommands maps no-arg CUI commands to Sultan methods.
var sultanNoArgCommands = cuiutil.NewCommandMap[usecase.SultanInteractorIF]().
	Add(usecase.SultanInteractorIF.Draw, "d", "draw").
	Add(usecase.SultanInteractorIF.Redeal, "rd", "redeal").
	Add(usecase.SultanInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.SultanInteractorIF.AutoComplete, "ac", "autocomplete").
	Add(usecase.SultanInteractorIF.Undo, "u", "undo").
	Add(usecase.SultanInteractorIF.Hint, "h", "hint").
	Add(usecase.SultanInteractorIF.ActionLog, "log", "l")

// sultanArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var sultanArgfulCommands = []string{"m", "move"}

// SultanCuiController スルタンCUIコントローラークラス
type SultanCuiController struct {
	si usecase.SultanInteractorIF
}

// NewSultanCuiController コンストラクタ
func NewSultanCuiController(si usecase.SultanInteractorIF) *SultanCuiController {
	return &SultanCuiController{si: si}
}

// Exec コマンド実行
func (c *SultanCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.si.Reset()
		},
		append(sultanNoArgCommands.Names(), sultanArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := sultanNoArgCommands.Lookup(cmd); ok {
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

// handleMove 移動コマンドを処理。
// 形式: m d <divanIdx> （ディヴァン→ファンデーション）、m w （ウェイスト→ファンデーション）。
// 省略形: m <divanIdx> （ディヴァン→ファンデーション）。
func (c *SultanCuiController) handleMove(args []string) string {
	if len(args) == 0 {
		return cuiutil.PromptRequest(i18n.T("sultan.promptSourceZone"), "m {0}")
	}
	// Shorthand: m <divanIdx>
	if idx, err := strconv.Atoi(args[0]); err == nil {
		return c.si.MoveDivanToFoundation(idx)
	}
	from := args[0]
	switch from {
	case "w":
		return c.si.MoveWasteToFoundation()
	case "d":
		if len(args) < 2 {
			return cuiutil.PromptRequest(i18n.T("sultan.promptDivanIndex"), "m d {0}")
		}
		idx, err := strconv.Atoi(args[1])
		if err != nil {
			return invalidArg("sultan.invalidDivanIndex", "val", args[1])
		}
		return c.si.MoveDivanToFoundation(idx)
	default:
		return invalidArg("sultan.invalidFromZone", "val", from)
	}
}
