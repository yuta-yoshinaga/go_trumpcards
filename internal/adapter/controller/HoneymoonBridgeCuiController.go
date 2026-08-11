//go:build !js || !wasm || solo

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HoneymoonBridgeCuiController ハネムーンブリッジCUIコントローラークラス
type HoneymoonBridgeCuiController struct {
	hi usecase.HoneymoonBridgeInteractorIF
}

// NewHoneymoonBridgeCuiController コンストラクタ
func NewHoneymoonBridgeCuiController(hi usecase.HoneymoonBridgeInteractorIF) *HoneymoonBridgeCuiController {
	return &HoneymoonBridgeCuiController{hi: hi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit          → ゲーム終了 ("bye.")
//	r / reset         → ゲームリセット (設定保持)
//	b / bid <n> <s>   → 契約を宣言する (n は 1..7、s は 0:NT 1:♠ 2:♣ 3:♥ 4:♦)
//	pass              → 競りを降りる
//	p / play <i>      → 手札の i 番目を出す
//	n / next          → 次のディールへ
//	g / giveup        → 投了
//	h / hint          → ヒント表示
//	log / l           → 棋譜表示
//
// **前半 13 トリックの引き合いにもコマンドは要らない。** play で進みます。
func (c *HoneymoonBridgeCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.hi.ResetWithConfig(c.hi.GetConfig()) },
		[]string{"b", "bid", "pass", "p", "play", "n", "next", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				return c.bid(args)
			case "pass":
				return c.hi.Pass(), true
			case "p", "play":
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax, c.hi.Play)
			case "n", "next":
				return c.hi.NextRound(), true
			case "g", "giveup":
				return c.hi.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.hi.Hint, c.hi.ActionLog)
			}
		},
	)
}

// bid はレベルとスートの 2 引数を取る。
//
// **どちらも既定値で埋めない。** スートの下限は 0 (ノートランプ) で、
// 「省略」ではなく明示的に選べる 5 つ目の選択肢です。
func (c *HoneymoonBridgeCuiController) bid(args []string) (string, bool) {
	level, errMsg, ok := cuiutil.ParseIntArg(args, "Level is required.", "Invalid level: %s.",
		1, domain.HoneymoonBridgeMaxLevel)
	if !ok {
		return errMsg, true
	}
	suit, errMsg, ok := cuiutil.ParseIntArg(args[1:], "Suit is required.", "Invalid suit: %s.",
		0, domain.CardDesignMax)
	if !ok {
		return errMsg, true
	}
	return c.hi.Bid(level, suit), true
}
