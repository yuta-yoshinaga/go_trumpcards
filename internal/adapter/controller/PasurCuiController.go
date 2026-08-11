//go:build !js || !wasm || extra

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PasurCuiController パスールCUIコントローラークラス
type PasurCuiController struct {
	pi usecase.PasurInteractorIF
}

// NewPasurCuiController コンストラクタ
func NewPasurCuiController(pi usecase.PasurInteractorIF) *PasurCuiController {
	return &PasurCuiController{pi: pi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit               → ゲーム終了 ("bye.")
//	r / reset              → ゲームリセット (設定保持)
//	p / play <i> [t...]    → 手札の i 番目を出す。t... は取る場札の番号（省略でトレール）
//	g / giveup             → 投了
//	h / hint               → ヒント表示
//	log / l                → 棋譜表示
//
// **取れるときにトレールはできません。** 取れる組み合わせがあるのに場札を
// 指定しないと、ドメインが弾きます。
func (c *PasurCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.pi.ResetWithConfig(c.pi.GetConfig()) },
		[]string{"p", "play", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return c.play(args)
			case "g", "giveup":
				return c.pi.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.pi.Hint, c.pi.ActionLog)
			}
		},
	)
}

// play は手札インデックスと、取る場札の番号（可変長）を取る。
func (c *PasurCuiController) play(args []string) (string, bool) {
	idx, errMsg, ok := cuiutil.ParseIntArg(args, "Card index is required.", "Invalid card index: %s.",
		cuiutil.NoMin, cuiutil.NoMax)
	if !ok {
		return errMsg, true
	}
	// **場札の指定は 0 個でもよい（トレール）。** 既定値では埋めない。
	table := make([]int, 0, len(args))
	for _, a := range args[1:] {
		v, err := strconv.Atoi(a)
		if err != nil {
			return "Invalid table index: " + a + ".", true
		}
		table = append(table, v)
	}
	return c.pi.Play(idx, table), true
}
