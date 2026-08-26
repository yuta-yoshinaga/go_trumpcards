//go:build !js || !wasm || classic

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// curdsAndWheyNoArgCommands maps no-arg CUI commands to interactor methods.
var curdsAndWheyNoArgCommands = cuiutil.NewCommandMap[usecase.CurdsAndWheyInteractorIF]().
	Add(usecase.CurdsAndWheyInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.CurdsAndWheyInteractorIF.Undo, "u", "undo").
	Add(usecase.CurdsAndWheyInteractorIF.Hint, "h", "hint").
	Add(usecase.CurdsAndWheyInteractorIF.ActionLog, "log", "l")

// CurdsAndWheyCuiController カーズ・アンド・ホエイのCUIコントローラークラス。
type CurdsAndWheyCuiController struct {
	si usecase.CurdsAndWheyInteractorIF
}

// NewCurdsAndWheyCuiController コンストラクタ。
func NewCurdsAndWheyCuiController(si usecase.CurdsAndWheyInteractorIF) *CurdsAndWheyCuiController {
	return &CurdsAndWheyCuiController{si: si}
}

// Exec コマンド実行。
// コマンド一覧:
//
//	r / reset                 新しいゲーム
//	m <from> <idx> <to>       列 from の idx 番目以降を列 to へ移す
//	u / undo                  直近の手を取り消す
//	g / giveup                投了
//	h / hint                  ヒント
//	log / l                   棋譜
func (c *CurdsAndWheyCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.si.Reset() },
		append(curdsAndWheyNoArgCommands.Names(), "m", "move"),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := curdsAndWheyNoArgCommands.Lookup(cmd); ok {
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

// handleMove `m <from> <idx> <to>` を処理する。
func (c *CurdsAndWheyCuiController) handleMove(args []string) string {
	if len(args) < 3 {
		return invalidArg("usageMFromCardindexTo")
	}
	from, err1 := strconv.Atoi(args[0])
	idx, err2 := strconv.Atoi(args[1])
	to, err3 := strconv.Atoi(args[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return invalidArg("invalidArgsMoveIntegers")
	}
	return c.si.MoveSequence(from, idx, to)
}
