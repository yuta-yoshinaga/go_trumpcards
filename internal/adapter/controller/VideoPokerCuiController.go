//go:build !js || !wasm || casino

package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// VideoPokerCuiController ビデオポーカーCUIコントローラークラス
type VideoPokerCuiController struct {
	vi usecase.VideoPokerInteractorIF
}

// NewVideoPokerCuiController コンストラクタ
func NewVideoPokerCuiController(vi usecase.VideoPokerInteractorIF) *VideoPokerCuiController {
	return &VideoPokerCuiController{
		vi: vi,
	}
}

// Exec ゲーム実行
// コマンド例: "r", "b 3", "h 0 1 4", "log", "q"
func (vpc *VideoPokerCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return vpc.vi.Reset() },
		[]string{"b", "bet", "h", "hold", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args, "betAmountRequired15", "invalidBetAmountPlain", 1, 5)
				if !ok {
					return errMsg, true
				}
				return vpc.vi.Bet(amount), true
			case "h", "hold":
				indices, errMsg := parseHoldIndices(args)
				if errMsg != "" {
					return errMsg, true
				}
				return vpc.vi.Hold(indices), true
			default:
				return handleCuiHintAndLog(cmd, vpc.vi.Hint, vpc.vi.ActionLog)
			}
		},
	)
}

// parseHoldIndices parses hold indices from command args.
// Empty args means hold no cards (returns []).
func parseHoldIndices(args []string) ([]int, string) {
	if len(args) == 0 {
		return []int{}, ""
	}
	indices := make([]int, 0, len(args))
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if arg == "" {
			continue
		}
		idx, err := strconv.Atoi(arg)
		if err != nil || idx < 0 || idx > 4 {
			return nil, invalidArg("invalidCardIndexHold")
		}
		indices = append(indices, idx)
	}
	return indices, ""
}
