//go:build !js || !wasm || extra2

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GuandanCuiController 掼蛋 (Guandan) CUIコントローラークラス
type GuandanCuiController struct {
	gi usecase.GuandanInteractorIF
}

// NewGuandanCuiController コンストラクタ
func NewGuandanCuiController(gi usecase.GuandanInteractorIF) *GuandanCuiController {
	return &GuandanCuiController{gi: gi}
}

// Exec コマンド実行
func (c *GuandanCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.gi.GetConfig()
			return c.gi.ResetWithConfig(cfg)
		},
		[]string{"p", "play", "ps", "pass", "t", "tribute", "n", "next", "ch", "check", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return guandanParsePlay(args, c.gi)
			case "ps", "pass":
				return c.gi.Pass(), true
			case "t", "tribute":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
					0, domain.GuandanHandSize-1, func(v int) string {
						return c.gi.ReturnTribute(v)
					})
			case "n", "next":
				return c.gi.NextHand(), true
			case "ch", "check":
				// **出さずに調べるだけ。**手札は動かない。
				idxs, skipped := cuiutil.ParseIntSlice(args)
				return cuiutil.PrependSkippedWarning(c.gi.Check(idxs), skipped), true
			default:
				return handleCuiLog(cmd, c.gi.ActionLog)
			}
		},
	)
}

// guandanParsePlay は `p <i> [<i> ...]` を解釈する。
//
// **役は複数枚で出す**ので、単一の添字では足りない。
func guandanParsePlay(args []string, gi usecase.GuandanInteractorIF) (string, bool) {
	if len(args) == 0 {
		return invalidArg("cardIndexesRequiredTriple"), true
	}
	idxs := make([]int, 0, len(args))
	seen := map[int]bool{}
	for _, a := range args {
		v, err := strconv.Atoi(a)
		if err != nil || v < 0 || v >= domain.GuandanHandSize {
			return invalidArg("invalidCardIndexRange", "val", a, "max", strconv.Itoa(domain.GuandanHandSize-1)), true
		}
		// **同じ札を 2 回数えられない。**
		if seen[v] {
			return "The same card was given twice: " + a + ".", true
		}
		seen[v] = true
		idxs = append(idxs, v)
	}
	return gi.PlayCards(idxs), true
}
