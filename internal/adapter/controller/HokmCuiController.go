//go:build !js || !wasm || classic

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HokmCuiController ホクムCUIコントローラークラス
type HokmCuiController struct {
	hi usecase.HokmInteractorIF
}

// NewHokmCuiController コンストラクタ
func NewHokmCuiController(hi usecase.HokmInteractorIF) *HokmCuiController {
	return &HokmCuiController{hi: hi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit       → ゲーム終了 ("bye.")
//	r / reset      → ゲームリセット (設定保持)
//	t / trump <s>  → 切り札スートを宣言する (1:♠ 2:♣ 3:♥ 4:♦、親のみ)
//	p / play <i>   → 手札の i 番目を出す
//	n / next       → 次のハンドへ
//	g / giveup     → 投了
//	h / hint       → ヒント表示
//	log / l        → 棋譜表示
func (c *HokmCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.hi.ResetWithConfig(c.hi.GetConfig()) },
		[]string{"t", "trump", "p", "play", "n", "next", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "t", "trump":
				return cuiutil.WithParsedInt(args, "Suit is required.", "Invalid suit: %s.",
					domain.CardDesignSpade, domain.CardDesignMax, c.hi.DeclareTrump)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.hi.Play)
			case "n", "next":
				return c.hi.NextHand(), true
			case "g", "giveup":
				return c.hi.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.hi.Hint, c.hi.ActionLog)
			}
		},
	)
}
