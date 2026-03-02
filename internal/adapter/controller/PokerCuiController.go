package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PokerCuiController ポーカーCUIコントローラークラス
type PokerCuiController struct {
	pi usecase.PokerInteractorIF
}

// NewPokerCuiController コンストラクタ
func NewPokerCuiController(pi usecase.PokerInteractorIF) *PokerCuiController {
	return &PokerCuiController{
		pi: pi,
	}
}

// Exec ゲーム実行
// コマンド例: "r", "e 0 2 4", "s", "b 20", "c", "ra 30", "f", "ck", "a", "q"
func (pcc *PokerCuiController) Exec(command string) string {
	res := ""
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "Unsupported command."
	}
	cmd := parts[0]
	switch cmd {
	case "q", "quit":
		res = "bye."
	case "r", "reset":
		res = pcc.pi.Reset()
	case "e", "exchange":
		indices := []int{}
		for _, p := range parts[1:] {
			idx, err := strconv.Atoi(p)
			if err == nil && 0 <= idx && idx <= 4 {
				indices = append(indices, idx)
			}
		}
		res = pcc.pi.Exchange(indices)
	case "s", "stand":
		res = pcc.pi.Stand()
	case "b", "bet":
		amount := 0
		if len(parts) > 1 {
			if a, err := strconv.Atoi(parts[1]); err == nil && a > 0 {
				amount = a
			}
		}
		res = pcc.pi.Action(domain.PokerActionBet, amount)
	case "c", "call":
		res = pcc.pi.Action(domain.PokerActionCall, 0)
	case "ra", "raise":
		amount := 0
		if len(parts) > 1 {
			if a, err := strconv.Atoi(parts[1]); err == nil && a > 0 {
				amount = a
			}
		}
		res = pcc.pi.Action(domain.PokerActionRaise, amount)
	case "f", "fold":
		res = pcc.pi.Action(domain.PokerActionFold, 0)
	case "ck", "check":
		res = pcc.pi.Action(domain.PokerActionCheck, 0)
	case "a", "allin":
		res = pcc.pi.Action(domain.PokerActionAllIn, 0)
	default:
		res = "Unsupported command."
	}
	return res
}
