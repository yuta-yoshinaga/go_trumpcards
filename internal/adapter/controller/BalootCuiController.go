//go:build !js || !wasm || extra2

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BalootCuiController バルートCUIコントローラークラス
type BalootCuiController struct {
	bi usecase.BalootInteractorIF
}

// NewBalootCuiController コンストラクタ
func NewBalootCuiController(bi usecase.BalootInteractorIF) *BalootCuiController {
	return &BalootCuiController{bi: bi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit        → ゲーム終了 ("bye.")
//	r / reset       → ゲームリセット (設定保持)
//	sun             → 切り札なし (Sun) を宣言する
//	hokom <suit>    → 指定スートを切り札として Hokom を宣言する (1:♠ 2:♣ 3:♥ 4:♦)
//	pass            → 宣言を見送る（親は見送れない）
//	p / play <i>    → 手札の i 番目を出す
//	n / next        → 次のラウンドへ
//	g / giveup      → 投了
//	h / hint        → ヒント表示
//	log / l         → 棋譜表示
func (c *BalootCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.bi.ResetWithConfig(c.bi.GetConfig()) },
		[]string{"sun", "hokom", "pass", "p", "play", "n", "next", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "sun":
				return c.bi.DeclareSun(), true
			case "hokom":
				return cuiutil.WithParsedInt(args, "Suit is required.", "Invalid suit: %s.",
					domain.CardDesignSpade, domain.CardDesignMax, c.bi.DeclareHokom)
			case "pass":
				return c.bi.PassDeclaration(), true
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.bi.Play)
			case "n", "next":
				return c.bi.NextRound(), true
			case "g", "giveup":
				return c.bi.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.bi.Hint, c.bi.ActionLog)
			}
		},
	)
}
