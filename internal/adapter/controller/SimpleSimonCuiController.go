//go:build !js || !wasm || classic

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// simpleSimonNoArgCommands maps no-arg CUI commands to interactor methods.
var simpleSimonNoArgCommands = cuiutil.NewCommandMap[usecase.SimpleSimonInteractorIF]().
	Add(usecase.SimpleSimonInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.SimpleSimonInteractorIF.Undo, "u", "undo").
	Add(usecase.SimpleSimonInteractorIF.Hint, "h", "hint").
	Add(usecase.SimpleSimonInteractorIF.ActionLog, "log", "l")

// SimpleSimonCuiController シンプル・サイモンのCUIコントローラークラス。
type SimpleSimonCuiController struct {
	si usecase.SimpleSimonInteractorIF
}

// NewSimpleSimonCuiController コンストラクタ。
func NewSimpleSimonCuiController(si usecase.SimpleSimonInteractorIF) *SimpleSimonCuiController {
	return &SimpleSimonCuiController{si: si}
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
func (c *SimpleSimonCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.si.Reset() },
		append(simpleSimonNoArgCommands.Names(), "m", "move"),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := simpleSimonNoArgCommands.Lookup(cmd); ok {
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
func (c *SimpleSimonCuiController) handleMove(args []string) string {
	if len(args) < 3 {
		return "Usage: m <from> <cardIndex> <to>."
	}
	from, err1 := strconv.Atoi(args[0])
	idx, err2 := strconv.Atoi(args[1])
	to, err3 := strconv.Atoi(args[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return invalidArg("invalidArgsMoveIntegers")
	}
	return c.si.MoveSequence(from, idx, to)
}
