//go:build !js || !wasm || extra2

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// doubleKlondikeNoArgCommands maps no-arg CUI commands to interactor methods.
var doubleKlondikeNoArgCommands = cuiutil.NewCommandMap[usecase.DoubleKlondikeInteractorIF]().
	Add(usecase.DoubleKlondikeInteractorIF.Draw, "d", "draw").
	Add(usecase.DoubleKlondikeInteractorIF.MoveWasteToFoundation, "mwf").
	Add(usecase.DoubleKlondikeInteractorIF.GiveUp, "g", "giveup").
	Add(usecase.DoubleKlondikeInteractorIF.AutoComplete, "ac", "autocomplete").
	Add(usecase.DoubleKlondikeInteractorIF.Undo, "u", "undo").
	Add(usecase.DoubleKlondikeInteractorIF.Hint, "h", "hint").
	Add(usecase.DoubleKlondikeInteractorIF.ActionLog, "log", "l")

// DoubleKlondikeCuiController ダブル・クロンダイクのCUIコントローラークラス。
type DoubleKlondikeCuiController struct {
	di usecase.DoubleKlondikeInteractorIF
}

// NewDoubleKlondikeCuiController コンストラクタ。
func NewDoubleKlondikeCuiController(di usecase.DoubleKlondikeInteractorIF) *DoubleKlondikeCuiController {
	return &DoubleKlondikeCuiController{di: di}
}

// Exec コマンド実行。
// コマンド一覧:
//
//	r / reset              新しいゲーム
//	d / draw               ストックからめくる
//	mwt <col>              ウェイスト→タブロー列
//	mwf                    ウェイスト→ファウンデーション
//	mtt <from> <idx> <to>  タブロー列→タブロー列
//	mtf <col>              タブロー列→ファウンデーション
//	ac / autocomplete      自動で出し切る
//	u / undo               直近の手を取り消す
//	g / giveup             投了
//	h / hint               ヒント
//	log / l                棋譜
func (c *DoubleKlondikeCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.di.Reset() },
		append(doubleKlondikeNoArgCommands.Names(), "mwt", "mtt", "mtf"),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := doubleKlondikeNoArgCommands.Lookup(cmd); ok {
				return fn(c.di), true
			}
			switch cmd {
			case "mwt":
				return c.withInts(args, 1, func(n []int) string { return c.di.MoveWasteToTableau(n[0]) }), true
			case "mtf":
				return c.withInts(args, 1, func(n []int) string { return c.di.MoveTableauToFoundation(n[0]) }), true
			case "mtt":
				return c.withInts(args, 3, func(n []int) string { return c.di.MoveTableauToTableau(n[0], n[1], n[2]) }), true
			default:
				return "", false
			}
		},
	)
}

// withInts parses the first `count` args as ints and calls fn.
func (c *DoubleKlondikeCuiController) withInts(args []string, count int, fn func([]int) string) string {
	if len(args) < count {
		return "Invalid arguments: expected " + strconv.Itoa(count) + " integer(s)."
	}
	nums := make([]int, count)
	for i := 0; i < count; i++ {
		n, err := strconv.Atoi(args[i])
		if err != nil {
			return "Invalid argument: " + args[i] + " is not an integer."
		}
		nums[i] = n
	}
	return fn(nums)
}
