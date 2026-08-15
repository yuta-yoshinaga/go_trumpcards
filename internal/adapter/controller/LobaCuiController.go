//go:build !js || !wasm || extra2

package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// LobaCuiController ロバCUIコントローラークラス
type LobaCuiController struct {
	li usecase.LobaInteractorIF
}

// NewLobaCuiController コンストラクタ
func NewLobaCuiController(li usecase.LobaInteractorIF) *LobaCuiController {
	return &LobaCuiController{li: li}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit          → ゲーム終了 ("bye.")
//	r / reset         → ゲームごとリセット (設定保持)
//	ds                → 山札から引く
//	dd                → 捨て札を取る
//	m <i,j,k>         → 手札の i,j,k をメルドとして出す
//	o <i> <m>         → 手札 i をメルド m に付ける
//	d <i>             → 手札 i を捨てる
//	n / next          → 次のラウンドへ進む
//	h / hint          → ヒント表示
//	log / l           → 棋譜表示
func (c *LobaCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.li.GetConfig()
			return c.li.ResetWithConfig(cfg)
		},
		[]string{"ds", "dd", "m", "meld", "o", "layoff", "d", "discard", "n", "next", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "ds", "drawstock":
				return c.li.DrawStock(), true
			case "dd", "drawdiscard":
				return c.li.DrawDiscard(), true
			case "m", "meld":
				return c.meld(args)
			case "o", "layoff":
				return c.layOff(args)
			case "d", "discard":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.li.Discard)
			case "n", "next":
				return c.li.NextRound(), true
			default:
				return handleCuiHintAndLog(cmd, c.li.Hint, c.li.ActionLog)
			}
		},
	)
}

// meld は `m 0,2,5` を解釈する。カンマ区切りなのは、複数枚を 1 コマンドで
// 指定する必要があるため。
func (c *LobaCuiController) meld(args []string) (string, bool) {
	if len(args) == 0 {
		return invalidArg("cardIndicesRequiredMeld"), true
	}
	parts := strings.Split(args[0], ",")
	idxs := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || n < 0 {
			return invalidArg("invalidCardIndex", "val", p), true
		}
		idxs = append(idxs, n)
	}
	return c.li.Meld(idxs), true
}

// layOff は `o <card> <meld>` を解釈する。
func (c *LobaCuiController) layOff(args []string) (string, bool) {
	if len(args) < 2 {
		return invalidArg("cardAndMeldIndexRequired"), true
	}
	card, err := strconv.Atoi(args[0])
	if err != nil || card < 0 {
		return invalidArg("invalidCardIndex", "val", args[0]), true
	}
	meld, err := strconv.Atoi(args[1])
	if err != nil || meld < 0 {
		return "Invalid meld index: " + args[1] + ".", true
	}
	return c.li.LayOff(card, meld), true
}
