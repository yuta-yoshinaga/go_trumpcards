//go:build !js || !wasm || classic

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BrusquembilleCuiController ブリュスカンビーユCUIコントローラークラス
type BrusquembilleCuiController struct {
	bi usecase.BrusquembilleInteractorIF
}

// NewBrusquembilleCuiController コンストラクタ
func NewBrusquembilleCuiController(bi usecase.BrusquembilleInteractorIF) *BrusquembilleCuiController {
	return &BrusquembilleCuiController{bi: bi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit         → ゲーム終了 ("bye.")
//	r / reset        → ゲームリセット (設定保持)
//	p / play <i>     → カードをプレイ
//	n / next         → 次のトリックへ
//	h / hint         → ヒント表示
//	log / l          → 棋譜表示
//	sp / setplayers <2-5> → 席数を変更して再開
func (c *BrusquembilleCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.bi.GetConfig()
			return c.bi.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "n", "next", "h", "hint", "log", "l", "sp", "setplayers",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.bi.Play)
			case "n", "next":
				return c.bi.NextTrick(), true
			case "sp", "setplayers":
				// **席数を変えられなければ 2〜5 人卓は誰にも届かない。**
				// ドメインが席数可変でも、入口が無ければ既定の 2 人卓しか始まらない。
				return cuiutil.WithParsedIntKeys(args, "playerCountRequired25", "invalidPlayerCount25",
					domain.BrusquembilleMinPlayerCnt, domain.BrusquembilleMaxPlayerCnt, func(v int) string {
						cfg := c.bi.GetConfig()
						cfg.PlayerCnt = v
						return c.bi.ResetWithConfig(cfg)
					})
			default:
				return handleCuiHintAndLog(cmd, c.bi.Hint, c.bi.ActionLog)
			}
		},
	)
}
