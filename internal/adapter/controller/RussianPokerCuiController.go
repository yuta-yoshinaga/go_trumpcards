//go:build !js || !wasm || casino

package controller

import (
	"math"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RussianPokerCuiController ロシアンポーカーCUIコントローラークラス
type RussianPokerCuiController struct {
	ri usecase.RussianPokerInteractorIF
}

// NewRussianPokerCuiController コンストラクタ
func NewRussianPokerCuiController(ri usecase.RussianPokerInteractorIF) *RussianPokerCuiController {
	return &RussianPokerCuiController{
		ri: ri,
	}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "e 0 2 4", "6", "sel 3", "p", "f", "force", "d", "q"
func (rc *RussianPokerCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return rc.ri.Reset() },
		[]string{"b", "bet", "e", "exchange", "6", "buy6th", "sel", "select", "p", "play", "f", "fold", "force", "fe", "d", "decline", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				ante, errMsg, ok := cuiutil.ParseIntArgKeys(args, "anteAmountRequired", "invalidAnteAmount", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return rc.ri.Bet(ante), true
			case "e", "exchange":
				indices, skipped := cuiutil.ParseBoundedIntSlice(args, 0, 4)
				// Refuse before playing. PrependSkippedWarning ran the move first and
				// put the warning above the new board, so a mistyped index was dropped
				// and the remaining ones played as a different, legal move (issue #5390).
				if len(skipped) > 0 {
					return invalidArg("invalidCardIndex", "val", strings.Join(skipped, ", ")), true
				}
				return rc.ri.Exchange(indices), true
			case "6", "buy6th":
				return rc.ri.Buy6th(), true
			case "sel", "select":
				idx, errMsg, ok := cuiutil.ParseIntArgKeys(args, "discardIndexRequired", "invalidIndexANumber", 0, 5)
				if !ok {
					return errMsg, true
				}
				return rc.ri.Select(idx), true
			case "p", "play":
				return rc.ri.Play(), true
			case "f", "fold":
				return rc.ri.Fold(), true
			case "force", "fe":
				return rc.ri.ForceExchange(), true
			case "d", "decline":
				return rc.ri.Decline(), true
			default:
				return handleCuiLog(cmd, rc.ri.ActionLog)
			}
		},
	)
}
