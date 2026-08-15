package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TrucoCuiController トゥルコCUIコントローラークラス
type TrucoCuiController struct {
	ti usecase.TrucoInteractorIF
}

// NewTrucoCuiController コンストラクタ
func NewTrucoCuiController(ti usecase.TrucoInteractorIF) *TrucoCuiController {
	return &TrucoCuiController{ti: ti}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit         → ゲーム終了 ("bye.")
//	r / reset        → マッチリセット (設定保持)
//	p / play <i>     → カードをプレイ
//	t / truco        → Truco を宣言 (または再引き上げ)
//	a / accept       → 宣言を受諾 (Quiero)
//	d / decline      → 宣言を拒否 (No Quiero)
//	n / next         → 次のバサ / マノへ
//	h / hint         → ヒント表示
//	log / l          → 棋譜表示
func (c *TrucoCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ti.GetConfig()
			return c.ti.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "t", "truco", "a", "accept", "d", "decline",
			"n", "next", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ti.Play)
			case "t", "truco":
				return c.ti.Truco(), true
			case "a", "accept":
				return c.ti.Respond(true), true
			case "d", "decline":
				return c.ti.Respond(false), true
			case "n", "next":
				return c.ti.Next(), true
			default:
				return handleCuiHintAndLog(cmd, c.ti.Hint, c.ti.ActionLog)
			}
		},
	)
}
