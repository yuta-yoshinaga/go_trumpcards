//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FreeBetBlackjackCuiController フリーベット・ブラックジャックCUIコントローラークラス
type FreeBetBlackjackCuiController struct {
	ci usecase.FreeBetBlackjackInteractorIF
}

// NewFreeBetBlackjackCuiController コンストラクタ
func NewFreeBetBlackjackCuiController(ci usecase.FreeBetBlackjackInteractorIF) *FreeBetBlackjackCuiController {
	return &FreeBetBlackjackCuiController{ci: ci}
}

// Exec ゲーム実行
// コマンド例: "r", "bet 50", "hit", "stand", "fd", "fs", "next", "hint", "log", "q"
//   - fd = 無料ダブル、fs = 無料スプリット（どちらもチップは不要）
func (cc *FreeBetBlackjackCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.ci.Reset() },
		[]string{"bet", "b", "hit", "h", "stand", "s", "fd", "freedouble",
			"fs", "freesplit", "next", "hint", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				ante, errMsg, ok := cuiutil.ParseIntArgKeys(args, "anteRequiredPlain", "invalidAnteANumber", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return cc.ci.PlaceBet(ante), true
			case "h", "hit":
				return cc.ci.Hit(), true
			case "s", "stand":
				return cc.ci.Stand(), true
			case "fd", "freedouble":
				return cc.ci.FreeDouble(), true
			case "fs", "freesplit":
				return cc.ci.FreeSplit(), true
			case "next":
				return cc.ci.NextRound(), true
			case "hint":
				return cc.ci.Hint(), true
			default:
				return handleCuiLog(cmd, cc.ci.ActionLog)
			}
		},
	)
}
