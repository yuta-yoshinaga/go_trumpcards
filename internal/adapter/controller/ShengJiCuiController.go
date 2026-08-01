//go:build !js || !wasm || classic

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ShengJiCuiController 升级 (Sheng Ji) CUIコントローラークラス
type ShengJiCuiController struct {
	gi usecase.ShengJiInteractorIF
}

// NewShengJiCuiController コンストラクタ
func NewShengJiCuiController(gi usecase.ShengJiInteractorIF) *ShengJiCuiController {
	return &ShengJiCuiController{gi: gi}
}

// Exec コマンド実行
func (c *ShengJiCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.gi.GetConfig()
			return c.gi.ResetWithConfig(cfg)
		},
		[]string{"d", "declare", "b", "bury", "p", "play", "n", "next", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "declare":
				// **0 はパス**なので下限は 0。
				return cuiutil.WithParsedInt(args, "Suit is required (0 to pass, 1-4 to declare).",
					"Invalid suit: %s.", domain.ShengJiNoTrump, domain.CardDesignDiamond, func(v int) string {
						return c.gi.Declare(v)
					})
			case "b", "bury":
				// **底牌を拾った直後だけ手札が 25 + 8 枚**あるので、上限が広い。
				return shengJiParseIndexes(args, domain.ShengJiKittySize,
					domain.ShengJiHandSize+domain.ShengJiKittySize-1, c.gi.BuryKitty)
			case "p", "play":
				// プレイ中は 25 枚。埋め戻し後の上限を使い回すと緩すぎる。
				return shengJiParseIndexes(args, 0, domain.ShengJiHandSize-1, c.gi.Play)
			case "n", "next":
				return c.gi.NextHand(), true
			default:
				return handleCuiLog(cmd, c.gi.ActionLog)
			}
		},
	)
}

// shengJiParseIndexes は `<cmd> <i> [<i> ...]` を解釈する。
//
// want が 0 でなければ枚数も検査する。**底牌はちょうど 8 枚**でなければならない。
// maxIdx は場面ごとに変わる: 底牌を拾った直後は 25 + 8 枚を持っているが、
// 埋め戻したあとは 25 枚しかない。
func shengJiParseIndexes(args []string, want, maxIdx int, apply func([]int) string) (string, bool) {
	if len(args) == 0 {
		return "Card indexes are required (e.g. p 0 1 for a pair).", true
	}
	if want > 0 && len(args) != want {
		return "Give exactly " + strconv.Itoa(want) + " card indexes.", true
	}
	idxs := make([]int, 0, len(args))
	seen := map[int]bool{}
	for _, a := range args {
		v, err := strconv.Atoi(a)
		if err != nil || v < 0 || v > maxIdx {
			return "Invalid card index: " + a + ". Please enter 0-" + strconv.Itoa(maxIdx) + ".", true
		}
		// **同じ札を 2 回数えられない。**通すと 1 枚から対子が作れてしまう。
		if seen[v] {
			return "The same card was given twice: " + a + ".", true
		}
		seen[v] = true
		idxs = append(idxs, v)
	}
	return apply(idxs), true
}
