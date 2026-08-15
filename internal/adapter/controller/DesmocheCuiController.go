//go:build !js || !wasm || extra3

package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DesmocheCuiController デスモチェCUIコントローラークラス
type DesmocheCuiController struct {
	di usecase.DesmocheInteractorIF
}

// NewDesmocheCuiController コンストラクタ
func NewDesmocheCuiController(di usecase.DesmocheInteractorIF) *DesmocheCuiController {
	return &DesmocheCuiController{di: di}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit          → ゲーム終了 ("bye.")
//	r / reset         → ゲームごとリセット (設定保持)
//	ds                → 山札から引く
//	dd                → 捨て札を取る
//	m <i,j,k>         → 手札の i,j,k をメルドとして出す
//	o <i> <m>         → 手札 i をメルド m に付ける
//	x <from> <i> <to> → メルド from の i 枚目をメルド to へ移す (desmoche)
//	d <i>             → 手札 i を捨てる
//	n / next          → 次のラウンドへ進む
//	h / hint          → ヒント表示
//	log / l           → 棋譜表示
func (c *DesmocheCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.di.GetConfig()
			return c.di.ResetWithConfig(cfg)
		},
		[]string{"ds", "dd", "m", "meld", "o", "layoff", "x", "desmoche", "d", "discard", "n", "next", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "ds", "drawstock":
				return c.di.DrawStock(), true
			case "dd", "drawdiscard":
				return c.di.DrawDiscard(), true
			case "m", "meld":
				return c.meld(args)
			case "o", "layoff":
				return c.layOff(args)
			case "x", "desmoche":
				return c.desmoche(args)
			case "d", "discard":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.di.Discard)
			case "n", "next":
				return c.di.NextRound(), true
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}

// meld は `m 0,2,5` を解釈する。カンマ区切りなのは、複数枚を 1 コマンドで
// 指定する必要があるため。
func (c *DesmocheCuiController) meld(args []string) (string, bool) {
	if len(args) == 0 {
		return "Card indices are required (for example: m 0,2,5).", true
	}
	parts := strings.Split(args[0], ",")
	idxs := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || n < 0 {
			return i18n.Tf("invalidCardIndex", "val", p), true
		}
		idxs = append(idxs, n)
	}
	return c.di.Meld(idxs), true
}

// layOff は `o <card> <meld>` を解釈する。
func (c *DesmocheCuiController) layOff(args []string) (string, bool) {
	if len(args) < 2 {
		return "Card index and meld index are required (for example: o 0 1).", true
	}
	card, ok := desmocheParseIdx(args[0])
	if !ok {
		return i18n.Tf("invalidCardIndex", "val", args[0]), true
	}
	meld, ok := desmocheParseIdx(args[1])
	if !ok {
		return "Invalid meld index: " + args[1] + ".", true
	}
	return c.di.LayOff(card, meld), true
}

// desmoche は `x <from> <card> <to>` を解釈する。
func (c *DesmocheCuiController) desmoche(args []string) (string, bool) {
	if len(args) < 3 {
		return "Source meld, card index and target meld are required (for example: x 0 2 1).", true
	}
	from, ok := desmocheParseIdx(args[0])
	if !ok {
		return "Invalid meld index: " + args[0] + ".", true
	}
	card, ok := desmocheParseIdx(args[1])
	if !ok {
		return i18n.Tf("invalidCardIndex", "val", args[1]), true
	}
	to, ok := desmocheParseIdx(args[2])
	if !ok {
		return "Invalid meld index: " + args[2] + ".", true
	}
	return c.di.Desmoche(from, card, to), true
}

// desmocheParseIdx は非負整数を解釈する。
func desmocheParseIdx(s string) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}
