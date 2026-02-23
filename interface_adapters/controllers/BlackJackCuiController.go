package controllers

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/usecases"
)

// BlackJackCuiController ブラックジャックCUIコントローラークラス
type BlackJackCuiController struct {
	bji usecases.BlackJackInteractorIF
}

// NewBlackJackCuiController コンストラクタ
func NewBlackJackCuiController(bji usecases.BlackJackInteractorIF) *BlackJackCuiController {
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
		amount := 0
		if len(parts) > 1 {
			amount, _ = strconv.Atoi(parts[1])
		}
		res = bcc.bji.Bet(amount)
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
