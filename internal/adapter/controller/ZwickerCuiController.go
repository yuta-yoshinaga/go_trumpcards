//go:build !js || !wasm || extra2

package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ZwickerCuiController ツヴィッカーCUIコントローラークラス
type ZwickerCuiController struct {
	zi usecase.ZwickerInteractorIF
}

// NewZwickerCuiController コンストラクタ
func NewZwickerCuiController(zi usecase.ZwickerInteractorIF) *ZwickerCuiController {
	return &ZwickerCuiController{zi: zi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit              → ゲーム終了 ("bye.")
//	r / reset             → ゲームごとリセット (設定保持)
//	t <i> <v> [t:a,b] [b:c] → 手札 i を値 v として使い、場札 a,b とビルド c を取る
//	b <i> <a,b> <v>       → 手札 i と場札 a,b で宣言値 v のビルドを作る
//	tr <i>                → 手札 i を場に置く
//	n / next              → 次のディールへ進む
//	h / hint              → ヒント表示
//	log / l               → 棋譜表示
func (c *ZwickerCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.zi.GetConfig()
			return c.zi.ResetWithConfig(cfg)
		},
		[]string{"t", "take", "b", "build", "tr", "trail", "n", "next", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "t", "take":
				return c.take(args)
			case "b", "build":
				return c.build(args)
			case "tr", "trail":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.zi.Trail)
			case "n", "next":
				return c.zi.NextRound(), true
			default:
				return handleCuiHintAndLog(cmd, c.zi.Hint, c.zi.ActionLog)
			}
		},
	)
}

// take は `t <hand> <value> [t:0,2] [b:1]` を解釈する。
//
// 値を明示させるのは、**A と絵札が 2 つの値を持つ**ため。札だけでは
// どちらのつもりか決まらない。
func (c *ZwickerCuiController) take(args []string) (string, bool) {
	if len(args) < 2 {
		return "Card index and value are required (for example: t 0 7 t:1,2).", true
	}
	hand, ok := zwickerParseIdx(args[0])
	if !ok {
		return invalidArg("invalidCardIndex", "val", args[0]), true
	}
	value, err := strconv.Atoi(strings.TrimSpace(args[1]))
	if err != nil || value <= 0 {
		return "Invalid value: " + args[1] + ".", true
	}
	var tableIdxs, buildIdxs []int
	for _, a := range args[2:] {
		list, prefix, ok := zwickerParseList(a)
		if !ok {
			return "Invalid selection: " + a + ".", true
		}
		if prefix == "b" {
			buildIdxs = append(buildIdxs, list...)
			continue
		}
		tableIdxs = append(tableIdxs, list...)
	}
	if len(tableIdxs) == 0 && len(buildIdxs) == 0 {
		return "Select something to capture (for example: t 0 7 t:1,2).", true
	}
	return c.zi.Take(hand, value, tableIdxs, buildIdxs), true
}

// build は `b <hand> <a,b> <value>` を解釈する。
func (c *ZwickerCuiController) build(args []string) (string, bool) {
	if len(args) < 3 {
		return "Card index, table cards and a value are required (for example: b 0 1,2 9).", true
	}
	hand, ok := zwickerParseIdx(args[0])
	if !ok {
		return invalidArg("invalidCardIndex", "val", args[0]), true
	}
	table, _, ok := zwickerParseList(args[1])
	if !ok || len(table) == 0 {
		return "Invalid table selection: " + args[1] + ".", true
	}
	value, err := strconv.Atoi(strings.TrimSpace(args[2]))
	if err != nil || value <= 0 {
		return "Invalid value: " + args[2] + ".", true
	}
	return c.zi.Build(hand, table, value), true
}

// zwickerParseIdx は非負整数を解釈する。
func zwickerParseIdx(s string) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

// zwickerParseList は `t:0,2` / `b:1` / `0,2` を添字集合と接頭辞に分ける。
func zwickerParseList(s string) ([]int, string, bool) {
	prefix := "t"
	body := strings.TrimSpace(s)
	if i := strings.Index(body, ":"); i >= 0 {
		prefix = strings.ToLower(strings.TrimSpace(body[:i]))
		body = body[i+1:]
		if prefix != "t" && prefix != "b" {
			return nil, "", false
		}
	}
	if body == "" {
		return nil, "", false
	}
	out := make([]int, 0, 4)
	for _, p := range strings.Split(body, ",") {
		n, ok := zwickerParseIdx(p)
		if !ok {
			return nil, "", false
		}
		out = append(out, n)
	}
	return out, prefix, true
}
