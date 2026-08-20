//go:build !js || !wasm || solo

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TeenDoPaanchCuiController 3-2-5 CUIコントローラークラス
type TeenDoPaanchCuiController struct {
	ti usecase.TeenDoPaanchInteractorIF
}

// NewTeenDoPaanchCuiController コンストラクタ
func NewTeenDoPaanchCuiController(ti usecase.TeenDoPaanchInteractorIF) *TeenDoPaanchCuiController {
	return &TeenDoPaanchCuiController{ti: ti}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit       → ゲーム終了 ("bye.")
//	r / reset      → ゲームリセット (設定保持)
//	t / trump <s>  → 切り札を宣言する (1:♠ 2:♣ 3:♥ 4:♦、ノルマ5の席のみ)
//	p / play <i>   → 手札の i 番目を出す
//	n / next       → 次のラウンドへ
//	g / giveup     → 投了
//	h / hint       → ヒント表示
//	log / l        → 棋譜表示
//
// **ノルマを宣言するコマンドは無い。** 3・2・5 は最初から割り当てられており、
// 選ぶ余地がありません。
func (c *TeenDoPaanchCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ti.ResetWithConfig(c.ti.GetConfig()) },
		[]string{"t", "trump", "p", "play", "n", "next", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "t", "trump":
				return cuiutil.WithParsedIntKeys(args, "suitRequired", "invalidSuit",
					domain.CardDesignSpade, domain.CardDesignDiamond, c.ti.DeclareTrump)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ti.Play)
			case "n", "next":
				return c.ti.NextRound(), true
			case "g", "giveup":
				return c.ti.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.ti.Hint, c.ti.ActionLog)
			}
		},
	)
}
