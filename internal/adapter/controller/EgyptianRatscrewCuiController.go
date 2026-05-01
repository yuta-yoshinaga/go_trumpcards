package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// EgyptianRatscrewCuiController エジプシャン・ラットスクリュー CUI コントローラー
type EgyptianRatscrewCuiController struct {
	ei usecase.EgyptianRatscrewInteractorIF
}

// NewEgyptianRatscrewCuiController コンストラクタ
func NewEgyptianRatscrewCuiController(ei usecase.EgyptianRatscrewInteractorIF) *EgyptianRatscrewCuiController {
	return &EgyptianRatscrewCuiController{ei: ei}
}

// Exec コマンド実行
//
//	q / quit        → ゲーム終了
//	r / reset       → ゲームリセット
//	s / step        → 現手番プレイヤーがカードをめくる
//	j / slap        → 人間プレイヤーがスラップを試みる
//	tick            → CPU の保留アクションを進める
//	log / l         → 棋譜表示
func (c *EgyptianRatscrewCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ei.Reset() },
		[]string{"s", "step", "j", "slap", "tick", "log", "l"},
		func(cmd string, _ []string) (string, bool) {
			switch cmd {
			case "s", "step":
				return c.ei.Step(), true
			case "j", "slap":
				return c.ei.Slap(0), true
			case "tick":
				return c.ei.Tick(), true
			case "log", "l":
				return c.ei.ActionLog(), true
			default:
				return "", false
			}
		},
	)
}
