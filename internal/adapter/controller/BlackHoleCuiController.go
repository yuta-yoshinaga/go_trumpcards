//go:build !js || !wasm || solo

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// blackHoleNoArgCommands maps no-arg CUI commands to interactor methods.
var blackHoleNoArgCommands = cuiutil.NewCommandMap[usecase.BlackHoleInteractorIF]().
	Add(usecase.BlackHoleInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.BlackHoleInteractorIF.Undo, "u", "undo").
	Add(usecase.BlackHoleInteractorIF.Hint, "h", "hint").
	Add(usecase.BlackHoleInteractorIF.ActionLog, "log", "l")

// BlackHoleCuiController ブラックホールのCUIコントローラークラス。
type BlackHoleCuiController struct {
	li usecase.BlackHoleInteractorIF
}

// NewBlackHoleCuiController コンストラクタ。
func NewBlackHoleCuiController(li usecase.BlackHoleInteractorIF) *BlackHoleCuiController {
	return &BlackHoleCuiController{li: li}
}

// Exec コマンド実行。
// コマンド一覧:
//
//	r / reset       新しいゲーム
//	m <fan>         扇 fan のトップをブラックホールへ積む
//	u / undo        直近の手を取り消す
//	g / giveup      投了
//	h / hint        ヒント
//	log / l         棋譜
func (c *BlackHoleCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.li.Reset() },
		append(blackHoleNoArgCommands.Names(), "m", "move"),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := blackHoleNoArgCommands.Lookup(cmd); ok {
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

// handleMove 移動コマンド `m <fan>` を処理する。
func (c *BlackHoleCuiController) handleMove(args []string) string {
	if len(args) < 1 {
		return invalidArg("usageMFan")
	}
	fan, err := strconv.Atoi(args[0])
	if err != nil {
		return "Invalid fan index: " + args[0] + "."
	}
	return c.li.MoveFanToBlackHole(fan)
}
