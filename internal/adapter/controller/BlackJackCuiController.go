package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BlackJackCuiController ブラックジャックCUIコントローラークラス
type BlackJackCuiController struct {
	bji usecase.BlackJackInteractorIF
}

// NewBlackJackCuiController コンストラクタ
func NewBlackJackCuiController(bji usecase.BlackJackInteractorIF) *BlackJackCuiController {
	return &BlackJackCuiController{
		bji: bji,
	}
}

// Exec ゲーム実行
// コマンド例: "r", "h", "s", "b 100", "d", "sp", "i", "di", "q"
func (bcc *BlackJackCuiController) Exec(command string) string {
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
		res = bcc.bji.Reset()
	case "h", "hit":
		res = bcc.bji.Hit()
	case "s", "stand":
		res = bcc.bji.Stand()
	case "b", "bet":
		if len(parts) < 2 {
			res = "Bet amount is required."
		} else {
			amount, err := strconv.Atoi(parts[1])
			if err != nil {
				res = "Invalid bet amount. Please enter a number."
			} else {
				res = bcc.bji.Bet(amount)
			}
		}
	case "d", "doubledown":
		res = bcc.bji.DoubleDown()
	case "sp", "split":
		res = bcc.bji.Split()
	case "i", "insurance":
		res = bcc.bji.Insurance()
	case "di", "declineinsurance":
		res = bcc.bji.DeclineInsurance()
	default:
		res = "Unsupported command."
	}
	return res
}
