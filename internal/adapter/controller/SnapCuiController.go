//go:build !js || !wasm || solo

package controller

import "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

// SnapCuiController スナップCUIコントローラークラス
type SnapCuiController struct {
	si usecase.SnapInteractorIF
}

// NewSnapCuiController コンストラクタ
func NewSnapCuiController(si usecase.SnapInteractorIF) *SnapCuiController {
	return &SnapCuiController{si: si}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit    → ゲーム終了 ("bye.")
//	r / reset   → ゲームリセット (設定保持)
//	s / step    → 1枚めくる
//	n / snap    → スナップを宣言する
//	t / tick    → CPU の保留アクションを進める
//	g / giveup  → 投了
//	h / hint    → ヒント表示
//	log / l     → 棋譜表示
//
// **宣言はいつでも打てます。** 手番でなくても、成立していなくても——
// 成立していなければペナルティになります。
func (c *SnapCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.si.ResetWithConfig(c.si.GetConfig()) },
		[]string{"s", "step", "n", "snap", "t", "tick", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, _ []string) (string, bool) {
			switch cmd {
			case "s", "step":
				return c.si.Step(), true
			case "n", "snap":
				return c.si.Snap(), true
			case "t", "tick":
				return c.si.Tick(), true
			case "g", "giveup":
				return c.si.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.si.Hint, c.si.ActionLog)
			}
		},
	)
}
