package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
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
//	sd <n>          → CPU 難易度 (0=Easy, 1=Normal, 2=Hard) を設定してリセット
func (c *SlapjackCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.si.Reset() },
		[]string{"s", "step", "j", "slap", "tick", "log", "l", "sd", "setdifficulty"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "s", "step":
				return c.si.Step(), true
			case "j", "slap":
				return c.si.Slap(0), true
			case "tick":
				return c.si.Tick(), true
			case "log", "l":
				return c.si.ActionLog(), true
			case "sd", "setdifficulty":
				// slapjack.helpSetDifficulty has advertised this since the game
				// shipped, but it was never wired up -- it fell through to
				// "unknown command" in both line and realtime mode while the Web
				// GUI had a difficulty selector. See issue #5179.
				return cuiutil.WithParsedInt(args,
					i18n.T("slapjack.difficultyRequired"),
					i18n.T("slapjack.difficultyInvalid"),
					int(domain.SlapjackCpuEasy), int(domain.SlapjackCpuHard),
					func(v int) string {
						cfg := c.si.GetConfig()
						cfg.CpuDifficulty = domain.SlapjackCpuDifficulty(v)
						return c.si.ResetWithConfig(cfg)
					})
			default:
				return "", false
			}
		},
	)
}
