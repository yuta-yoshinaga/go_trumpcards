//go:build !js || !wasm || extra2

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ShelemCuiController シェレムCUIコントローラークラス
type ShelemCuiController struct {
	si usecase.ShelemInteractorIF
}

// NewShelemCuiController コンストラクタ
func NewShelemCuiController(si usecase.ShelemInteractorIF) *ShelemCuiController {
	return &ShelemCuiController{si: si}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                   → ゲーム終了 ("bye.")
//	r / reset                  → ゲームリセット (設定保持)
//	b / bid <n>                → n 点で入札する (100..165、5刻み)
//	shelem                     → Shelem（全トリック独占）を宣言する
//	pass                       → 競りを降りる
//	d / discard <i> <i> <i> <i> <suit> → 4枚捨てて切り札を決める (1:♠ 2:♣ 3:♥ 4:♦)
//	p / play <i>               → 手札の i 番目を出す
//	n / next                   → 次のラウンドへ
//	g / giveup                 → 投了
//	h / hint                   → ヒント表示
//	log / l                    → 棋譜表示
func (c *ShelemCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.si.ResetWithConfig(c.si.GetConfig()) },
		[]string{"b", "bid", "shelem", "pass", "d", "discard", "p", "play", "n", "next", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				return cuiutil.WithParsedIntKeys(args, "bidRequired", "invalidBid",
					domain.ShelemMinBid, domain.ShelemMaxBid, c.si.Bid)
			case "shelem":
				return c.si.BidShelem(), true
			case "pass":
				return c.si.Pass(), true
			case "d", "discard":
				return c.discard(args)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.si.Play)
			case "n", "next":
				return c.si.NextRound(), true
			case "g", "giveup":
				return c.si.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.si.Hint, c.si.ActionLog)
			}
		},
	)
}

// discard 捨て札は 4 つのインデックス + 切り札スートの 5 引数を取る。
//
// **どれも既定値で埋めない。** 埋めると捨てていない札が捨てられたり、
// 選んでいないスートが切り札になる。
func (c *ShelemCuiController) discard(args []string) (string, bool) {
	if len(args) < domain.ShelemWidowSize+1 {
		return invalidArg("fourIndicesAndSuitRequired"), true
	}
	indices := make([]int, 0, domain.ShelemWidowSize)
	for i := range domain.ShelemWidowSize {
		v, errMsg, ok := cuiutil.ParseIntArgKeys(args[i:], "cardIndexRequired", "invalidCardIndex",
			cuiutil.NoMin, cuiutil.NoMax)
		if !ok {
			return errMsg, true
		}
		indices = append(indices, v)
	}
	suit, errMsg, ok := cuiutil.ParseIntArgKeys(args[domain.ShelemWidowSize:], "suitRequired", "invalidSuit",
		domain.CardDesignSpade, domain.CardDesignMax)
	if !ok {
		return errMsg, true
	}
	return c.si.Discard(indices, suit), true
}
