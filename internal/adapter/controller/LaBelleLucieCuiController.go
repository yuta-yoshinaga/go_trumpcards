//go:build !js || !wasm || classic

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// laBelleLucieNoArgCommands maps no-arg CUI commands to interactor methods.
var laBelleLucieNoArgCommands = cuiutil.NewCommandMap[usecase.LaBelleLucieInteractorIF]().
	Add(usecase.LaBelleLucieInteractorIF.Redeal, "rd", "redeal").
	Add(usecase.LaBelleLucieInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.LaBelleLucieInteractorIF.AutoComplete, "ac", "autocomplete").
	Add(usecase.LaBelleLucieInteractorIF.Undo, "u", "undo").
	Add(usecase.LaBelleLucieInteractorIF.Hint, "h", "hint").
	Add(usecase.LaBelleLucieInteractorIF.ActionLog, "log", "l")

// LaBelleLucieCuiController ラ・ベル・ルーシーのCUIコントローラークラス。
type LaBelleLucieCuiController struct {
	li usecase.LaBelleLucieInteractorIF
}

// NewLaBelleLucieCuiController コンストラクタ。
func NewLaBelleLucieCuiController(li usecase.LaBelleLucieInteractorIF) *LaBelleLucieCuiController {
	return &LaBelleLucieCuiController{li: li}
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
func (c *LaBelleLucieCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.li.Reset() },
		append(laBelleLucieNoArgCommands.Names(), "m", "move"),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := laBelleLucieNoArgCommands.Lookup(cmd); ok {
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
func (c *LaBelleLucieCuiController) handleMove(args []string) string {
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
