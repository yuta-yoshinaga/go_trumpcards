//go:build !js || !wasm || extra3

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MinibridgeCuiController ミニブリッジCUIコントローラークラス
type MinibridgeCuiController struct {
	mi usecase.MinibridgeInteractorIF
}

// NewMinibridgeCuiController コンストラクタ
func NewMinibridgeCuiController(mi usecase.MinibridgeInteractorIF) *MinibridgeCuiController {
	return &MinibridgeCuiController{mi: mi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit             → ゲーム終了 ("bye.")
//	r / reset            → ゲームリセット (設定保持)
//	c / contract <n> <s> → 契約を選ぶ (n は 1..7、s は 0:NT 1:♠ 2:♣ 3:♥ 4:♦)
//	p / play <i>         → 手札の i 番目を出す（ダミーの手番も自分で出す）
//	n / next             → 次のディールへ
//	g / giveup           → 投了
//	h / hint             → ヒント表示
//	log / l              → 棋譜表示
//
// **競りはありません。** HCP を全員が公開申告し、合計の高いペアのうち多い方が
// デクレアラーになります。宣言するのは契約だけです。
func (c *MinibridgeCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.mi.ResetWithConfig(c.mi.GetConfig()) },
		[]string{"c", "contract", "p", "play", "n", "next", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "c", "contract":
				return c.contract(args)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.mi.Play)
			case "n", "next":
				return c.mi.NextRound(), true
			case "g", "giveup":
				return c.mi.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.mi.Hint, c.mi.ActionLog)
			}
		},
	)
}

// contract はレベルと種別の 2 引数を取る。
//
// **どちらも既定値で埋めない。** スートの下限は 0 (ノートランプ) で、
// 「省略」ではなく明示的に選べる 5 つ目の選択肢です。
func (c *MinibridgeCuiController) contract(args []string) (string, bool) {
	level, errMsg, ok := cuiutil.ParseIntArg(args, "Level is required.", "Invalid level: %s.",
		1, domain.MinibridgeMaxLevel)
	if !ok {
		return errMsg, true
	}
	suit, errMsg, ok := cuiutil.ParseIntArg(args[1:], "Suit is required.", "Invalid suit: %s.",
		0, domain.CardDesignMax)
	if !ok {
		return errMsg, true
	}
	return c.mi.Contract(level, suit), true
}
