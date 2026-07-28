//go:build !js || !wasm || extra2

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KempsCuiController はケムプスの CUI コントローラー。
type KempsCuiController struct {
	ki usecase.KempsInteractorIF
}

// NewKempsCuiController はコンストラクタ。
func NewKempsCuiController(ki usecase.KempsInteractorIF) *KempsCuiController {
	return &KempsCuiController{ki: ki}
}

// Exec はコマンドを実行する。
//
//	q / quit              → ゲーム終了
//	r / reset             → ゲームリセット
//	s <h> <f> / swap <h> <f> → 手札 h をフィールド f と交換する
//	p / pass              → 交換を見送る (宣言フェーズでは宣言を見送る)
//	sig <n> / signal <n>  → シグナル種別 (0=Sound, 1=Blink) を設定する
//	k / kemps             → Kemps を宣言する
//	c <seat> / counter <seat> → 相手 seat に Counter-Kemps を宣言する
//	n / next              → 次のラウンドへ進む
//	log / l               → 棋譜表示
func (c *KempsCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ki.Reset() },
		[]string{"s", "swap", "p", "pass", "sig", "signal", "k", "kemps", "c", "counter", "n", "next", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "s", "swap":
				h, f := kempsArgInt(args, 0), kempsArgInt(args, 1)
				return c.ki.Swap(h, f), true
			case "p", "pass":
				return c.ki.Pass(), true
			case "sig", "signal":
				return c.ki.SetSignal(kempsArgInt(args, 0)), true
			case "k", "kemps":
				return c.ki.DeclareKemps(), true
			case "c", "counter":
				return c.ki.DeclareCounterKemps(kempsArgInt(args, 0)), true
			case "n", "next":
				return c.ki.NextRound(), true
			case "log", "l":
				return c.ki.ActionLog(), true
			default:
				return "", false
			}
		},
	)
}

// kempsArgInt は args[i] を整数として取り出す (存在しない/不正なら 0)。
func kempsArgInt(args []string, i int) int {
	if i < 0 || i >= len(args) {
		return 0
	}
	if v, err := strconv.Atoi(args[i]); err == nil {
		return v
	}
	return 0
}
