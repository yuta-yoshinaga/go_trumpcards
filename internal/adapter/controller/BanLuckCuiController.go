//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BanLuckCuiController バンラックCUIコントローラークラス
type BanLuckCuiController struct {
	ci usecase.BanLuckInteractorIF
}

// NewBanLuckCuiController コンストラクタ
func NewBanLuckCuiController(ci usecase.BanLuckInteractorIF) *BanLuckCuiController {
	return &BanLuckCuiController{ci: ci}
}

// Exec ゲーム実行
// コマンド例: "r", "bet 50", "hit", "stand", "next", "hint", "log", "q"
//   - 親のラウンドは "bet 0" (親は賭けない)
func (cc *BanLuckCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return cc.ci.Reset() },
		[]string{"bet", "b", "hit", "h", "stand", "s", "next", "hint", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				// **0 を弾かない。** 親のラウンドでは 0 が正しい入力。
				bet, errMsg, ok := cuiutil.ParseIntArgKeys(args, "betRequired", "invalidBetANumber", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return cc.ci.PlaceBet(bet), true
			case "h", "hit":
				return cc.ci.Hit(), true
			case "s", "stand":
				return cc.ci.Stand(), true
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
