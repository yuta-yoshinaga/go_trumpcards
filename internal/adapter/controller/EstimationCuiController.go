//go:build !js || !wasm || classic

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// EstimationCuiController エスティメーションCUIコントローラークラス
type EstimationCuiController struct {
	ei usecase.EstimationInteractorIF
}

// NewEstimationCuiController コンストラクタ
func NewEstimationCuiController(ei usecase.EstimationInteractorIF) *EstimationCuiController {
	return &EstimationCuiController{ei: ei}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit       → ゲーム終了 ("bye.")
//	r / reset      → ゲームリセット (設定保持)
//	t / trump <s>  → 切り札スートを決める (1:♠ 2:♣ 3:♥ 4:♦、親のみ)
//	b / bid <n>    → 獲得予定トリック数を宣言する (0..13、0 は Dash Call)
//	p / play <i>   → 手札の i 番目を出す
//	n / next       → 次のラウンドへ
//	g / giveup     → 投了
//	h / hint       → ヒント表示
//	log / l        → 棋譜表示
func (c *EstimationCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ei.ResetWithConfig(c.ei.GetConfig()) },
		[]string{"t", "trump", "b", "bid", "p", "play", "n", "next", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "t", "trump":
				return cuiutil.WithParsedInt(args, "Suit is required.", "Invalid suit: %s.",
					domain.CardDesignSpade, domain.CardDesignMax, c.ei.SelectTrump)
			case "b", "bid":
				return cuiutil.WithParsedInt(args, "Bid is required.", "Invalid bid: %s.",
					0, domain.EstimationHandSize, c.ei.Bid)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ei.Play)
			case "n", "next":
				return c.ei.NextRound(), true
			case "g", "giveup":
				return c.ei.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.ei.Hint, c.ei.ActionLog)
			}
		},
	)
}
