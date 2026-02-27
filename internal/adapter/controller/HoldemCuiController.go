package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HoldemCuiController テキサスホールデムCUIコントローラークラス
type HoldemCuiController struct {
	hi usecase.HoldemInteractorIF
}

// NewHoldemCuiController コンストラクタ
func NewHoldemCuiController(hi usecase.HoldemInteractorIF) *HoldemCuiController {
	return &HoldemCuiController{hi: hi}
}

// Exec コマンド実行
func (c *HoldemCuiController) Exec(command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return "コマンドが不明です: " + command
	}
	switch fields[0] {
	case "q", "quit":
		return "bye."
	case "r", "reset":
		return c.hi.Reset()
	case "f", "fold":
		return c.hi.Action(domain.HoldemActionFold, 0)
	case "ck", "check":
		return c.hi.Action(domain.HoldemActionCheck, 0)
	case "c", "call":
		return c.hi.Action(domain.HoldemActionCall, 0)
	case "b", "bet":
		amount := 0
		if len(fields) > 1 {
			if parsed, err := strconv.Atoi(fields[1]); err == nil {
				amount = parsed
			}
		}
		return c.hi.Action(domain.HoldemActionBet, amount)
	case "ra", "raise":
		amount := 0
		if len(fields) > 1 {
			if parsed, err := strconv.Atoi(fields[1]); err == nil {
				amount = parsed
			}
		}
		return c.hi.Action(domain.HoldemActionRaise, amount)
	case "a", "allin":
		return c.hi.Action(domain.HoldemActionAllIn, 0)
	default:
		return "コマンドが不明です: " + command
	}
}
