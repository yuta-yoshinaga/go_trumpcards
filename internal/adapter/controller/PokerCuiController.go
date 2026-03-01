package controller

import (
	"strconv"
	"strings"

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
// コマンド例: "r", "e 0 2 4", "s", "b 20", "c", "ra 30", "f", "ck", "q"
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
		// 交換するカードのインデックスをパース (例: "e 0 1 2")
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
		// Negative/zero amounts fall back to 0 (no bet). Unlike BlackJack/Holdem,
		// Poker's Bet(0) is a valid "no bet" action so we silently default.
		amount := 0
		if len(parts) > 1 {
			if a, err := strconv.Atoi(parts[1]); err == nil && a > 0 {
				amount = a
			}
		}
		res = pcc.pi.Bet(amount)
	case "c", "call":
		res = pcc.pi.Call()
	case "ra", "raise":
		// Same fallback-to-zero semantics as "bet" above.
		amount := 0
		if len(parts) > 1 {
			if a, err := strconv.Atoi(parts[1]); err == nil && a > 0 {
				amount = a
			}
		}
		res = pcc.pi.Raise(amount)
	case "f", "fold":
		res = pcc.pi.Fold()
	case "ck", "check":
		res = pcc.pi.Check()
	default:
		res = "Unsupported command."
	}
	return res
}
