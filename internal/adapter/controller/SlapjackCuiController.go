package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SlapjackCuiController スラップジャック CUI コントローラー
type SlapjackCuiController struct {
	si usecase.SlapjackInteractorIF
}

// NewSlapjackCuiController コンストラクタ
func NewSlapjackCuiController(si usecase.SlapjackInteractorIF) *SlapjackCuiController {
	return &SlapjackCuiController{si: si}
}

// Exec コマンド実行
//
//	q / quit        → ゲーム終了
//	r / reset       → ゲームリセット
//	s / step        → 現手番プレイヤーがカードをめくる
//	j / slap        → 人間プレイヤーがスラップを試みる
//	tick            → CPU の保留アクションを進める
//	log / l         → 棋譜表示
func (c *SlapjackCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.si.Reset() },
		[]string{"s", "step", "j", "slap", "tick", "log", "l"},
		func(cmd string, _ []string) (string, bool) {
			switch cmd {
			case "s", "step":
				return c.si.Step(), true
			case "j", "slap":
				return c.si.Slap(0), true
			case "tick":
				return c.si.Tick(), true
			case "log", "l":
				return c.si.ActionLog(), true
			default:
				return "", false
			}
		},
	)
}
