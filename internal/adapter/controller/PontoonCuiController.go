//go:build !js || !wasm || extra2

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PontoonCuiController ポンツーン CUI コントローラークラス
type PontoonCuiController struct {
	pi usecase.PontoonInteractorIF
}

// NewPontoonCuiController コンストラクタ
func NewPontoonCuiController(pi usecase.PontoonInteractorIF) *PontoonCuiController {
	return &PontoonCuiController{pi: pi}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "s", "t", "buy 50", "sp", "bt", "bs", "log", "q"
func (pc *PontoonCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return pc.pi.Reset() },
		[]string{"b", "bet", "deal", "s", "stick", "t", "twist", "buy", "sp", "split", "bt", "bs", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args,
					"betAmountRequired", "invalidBetAmount",
					domain.PontoonMinBet, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return pc.pi.Bet(amount), true
			case "deal":
				return pc.pi.Deal(), true
			case "s", "stick":
				return pc.pi.Stick(), true
			case "t", "twist":
				return pc.pi.Twist(), true
			case "buy":
				extra, errMsg, ok := cuiutil.ParseIntArgKeys(args, "extraStakeRequired", "invalidStakeANumber",
					domain.PontoonMinBet, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return pc.pi.Buy(extra), true
			case "sp", "split":
				return pc.pi.Split(), true
			case "bt":
				return pc.pi.BankerTwist(), true
			case "bs":
				return pc.pi.BankerStay(), true
			default:
				if cmd == "log" || cmd == "l" {
					return pc.pi.ActionLog(), true
				}
				return "", false
			}
		},
	)
}
