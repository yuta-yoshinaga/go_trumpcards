//go:build !js || !wasm || classic

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// shamrocksNoArgCommands maps no-arg CUI commands to interactor methods.
var shamrocksNoArgCommands = cuiutil.NewCommandMap[usecase.ShamrocksInteractorIF]().
	Add(usecase.ShamrocksInteractorIF.Redeal, "rd", "redeal").
	Add(usecase.ShamrocksInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.ShamrocksInteractorIF.AutoComplete, "ac", "autocomplete").
	Add(usecase.ShamrocksInteractorIF.Undo, "u", "undo").
	Add(usecase.ShamrocksInteractorIF.Hint, "h", "hint").
	Add(usecase.ShamrocksInteractorIF.ActionLog, "log", "l")

// ShamrocksCuiController シャムロックスのCUIコントローラークラス。
type ShamrocksCuiController struct {
	li usecase.ShamrocksInteractorIF
}

// NewShamrocksCuiController コンストラクタ。
func NewShamrocksCuiController(li usecase.ShamrocksInteractorIF) *ShamrocksCuiController {
	return &ShamrocksCuiController{li: li}
}

// Exec コマンド実行。
// コマンド一覧:
//
//	r / reset            新しいゲーム
//	m <from> <to>        扇 from のトップを扇 to へ
//	m <from> f           扇 from のトップをファウンデーションへ
//	rd / redeal          集めてシャッフルし配り直す
//	ac / autocomplete    出せる札を自動で出し切る
//	u / undo             直近の手を取り消す
//	g / giveup           投了
//	h / hint             ヒント
//	log / l              棋譜
func (c *ShamrocksCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.li.Reset() },
		append(shamrocksNoArgCommands.Names(), "m", "move"),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := shamrocksNoArgCommands.Lookup(cmd); ok {
				return fn(c.li), true
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

// handleMove 移動コマンド `m <from> <to|f>` を処理する。
func (c *ShamrocksCuiController) handleMove(args []string) string {
	if len(args) < 2 {
		return invalidArg("usageMFromToF")
	}
	from, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidSourceFanDot", "val", args[0])
	}
	if args[1] == "f" || args[1] == "F" {
		return c.li.MoveFanToFoundation(from)
	}
	to, err := strconv.Atoi(args[1])
	if err != nil {
		return invalidArg("invalidDestinationFanOrF", "val", args[1])
	}
	return c.li.MoveFanToFan(from, to)
}
