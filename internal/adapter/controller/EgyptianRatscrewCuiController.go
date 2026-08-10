package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
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
//	sd <n>          → CPU 難易度 (0=Easy, 1=Normal, 2=Hard) を設定してリセット
func (c *EgyptianRatscrewCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ei.Reset() },
		[]string{"s", "step", "j", "slap", "tick", "log", "l", "sd", "setdifficulty"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "s", "step":
				return c.ei.Step(), true
			case "j", "slap":
				return c.ei.Slap(0), true
			case "tick":
				return c.ei.Tick(), true
			case "log", "l":
				return c.ei.ActionLog(), true
			case "sd", "setdifficulty":
				// Same gap as Slapjack: advertised by helpSetDifficulty, never
				// implemented. See issue #5179.
				return cuiutil.WithParsedInt(args,
					i18n.T("egyptianratscrew.difficultyRequired"),
					i18n.T("egyptianratscrew.difficultyInvalid"),
					int(domain.EgyptianRatscrewCpuEasy), int(domain.EgyptianRatscrewCpuHard),
					func(v int) string {
						cfg := c.ei.GetConfig()
						cfg.CpuDifficulty = domain.EgyptianRatscrewCpuDifficulty(v)
						return c.ei.ResetWithConfig(cfg)
					})
			default:
				return "", false
			}
		},
	)
}
