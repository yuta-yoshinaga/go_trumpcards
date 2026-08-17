package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PigsTailCuiController ぶたのしっぽCUIコントローラークラス
type PigsTailCuiController struct {
	pti usecase.PigsTailInteractorIF
}

// NewPigsTailCuiController コンストラクタ
func NewPigsTailCuiController(pti usecase.PigsTailInteractorIF) *PigsTailCuiController {
	return &PigsTailCuiController{pti: pti}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit              → ゲーム終了 ("bye.")
//	r / reset             → ゲームリセット (設定保持)
//	d / draw              → 輪から1枚引く
//	sp / setplayers <2-6> → 参加人数を変更してリセット
//	log / l               → 行動ログ表示
func (c *PigsTailCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.pti.Reset(c.pti.GetConfig()) },
		[]string{"d", "draw", "sp", "setplayers", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "draw":
				return c.pti.Action(0), true
			case "sp", "setplayers":
				// Web は 2〜6 人を選べるのに CUI からは変えられず、常に既定の
				// 4 人で始まっていた (#5521)。他の設定は現在値から引き継ぐ。
				return cuiutil.WithParsedIntKeys(args, "playerCountRequired26", "invalidPlayerCount26",
					domain.PigsTailMinPlayers, domain.PigsTailMaxPlayers, func(v int) string {
						cfg := c.pti.GetConfig()
						cfg.PlayerCount = v
						return c.pti.Reset(cfg)
					})
			default:
				return handleCuiLog(cmd, c.pti.ActionLog)
			}
		},
	)
}
